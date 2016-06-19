package pft

import (
    "net"
    "fmt"
    "time"
    "os"
    "log"
    "strings"
    "bufio"
    "bytes"
    "path/filepath"
    "strconv"
)

type InboundPacket struct {
    sender *net.UDPAddr
    data   []byte
    size   int
}

type RemoteClient struct {
    addr                 *net.UDPAddr

    download_deadline    time.Time // if this deadline is passed, resend last packet
    download_state       int
    download_rid         string    // for when state == HALF_OPEN: REQ or it's answer may get lost, so we need this for timeouts
    download             *Download // this is for ACK'ed downloads, do not put download_state and download_rid into Download struct

    upload_state         int
    upload_rid           string
    upload_original_hash []byte
    upload_file_path     string
    upload_file_size     uint64

    push_rid             string
    push_deadline        time.Time

    cntf_deadline        time.Time
    cntf_state           int
    cntf_rid             string
}

type Peer struct {
    remote_filter *net.UDPAddr // if remote_filter != nil: only accept packets from this address (client mode)
    conn          *net.UDPConn
    read_chan     chan InboundPacket
    remotes       map[string]*RemoteClient
    torrent_map   map[string]Torrent
}

func MakePeer(localAddr *net.UDPAddr, remoteAddr *net.UDPAddr) Peer {
    conn, err := net.ListenUDP("udp", localAddr)
    CheckError(err)

    return Peer{
        remote_filter: remoteAddr,
        conn: conn,
        read_chan: make(chan InboundPacket),
        remotes: make(map[string]*RemoteClient),
    }
}

func (this *Peer) SetTorrentMap(torrent_map map[string]Torrent) {
    this.torrent_map = torrent_map
}

func (this *Peer) GetRemote(remote_addr *net.UDPAddr) *RemoteClient {
    remote, ok := this.remotes[remote_addr.String()]
    if !ok {
        remote = new(RemoteClient)
        remote.download_state = CLOSED
        remote.upload_state = CLOSED
        remote.addr = remote_addr
        this.remotes[remote_addr.String()] = remote
    }
    return remote
}

func (this *Peer) HandleReq(remote *RemoteClient, rid string) {
    rid = filepath.FromSlash(rid)
    remote.upload_rid = rid
    remote.upload_state = CLOSED

    if strings.HasPrefix(rid, "file:") {
        upload_file_path := filepath.Clean(filepath.Join(GetFileDir(), rid[5:]))
        if !strings.HasPrefix(upload_file_path, GetFileDir()) {
            log.Println("Client trying to access files outside file dir")
            this.conn.WriteToUDP(EncodeReqNack(), remote.addr)
            return
        }

        f, err := os.Open(upload_file_path)
        stat, err := f.Stat()
        if err != nil {
            log.Println("Client trying to access inaccessbile file", upload_file_path, "error is", err.Error())
            this.conn.WriteToUDP(EncodeReqNack(), remote.addr)
            return
        }

        hash := GetFileHash(upload_file_path)
        remote.upload_original_hash = hash
        remote.upload_file_path = upload_file_path
        remote.upload_file_size = uint64(stat.Size())

        req_ack := EncodeReqAck(uint64(stat.Size()), hash)
        this.conn.WriteToUDP(req_ack, remote.addr)
        remote.upload_state = OPEN
        log.Println("serving file", upload_file_path)
    } else if rid == "file-list" {
        size, hash := GetFileListSizeAndHash(GetFileDir())
        req_ack := EncodeReqAck(size, hash)
        this.conn.WriteToUDP(req_ack, remote.addr)
        remote.upload_state = OPEN
        remote.upload_original_hash = hash
        log.Println("serving file-list")
    } else {
        this.conn.WriteToUDP(EncodeReqNack(), remote.addr)
    }
}

func (this *Peer) HandlePush(remote *RemoteClient, rid string) {
    rid = filepath.FromSlash(rid)
    log.Println("got push for rid:" + rid)
    this.conn.WriteToUDP(EncodePushAck(), remote.addr)
    log.Println("sent PUSH-ACK")

    this.Download(rid, remote.addr)
}

func (this *Peer) DownloadNextBlock(remote *RemoteClient) {
    if remote.download.IsFinished() {
        remote.download.Close()
        if (remote.download_rid == "file-list") {
            printFileList()
        }
        remote.download_state = CLOSED
    } else {
        this.conn.WriteToUDP(remote.download.CreateNextGet(), remote.addr)
    }
}

func (this *Peer) HandleReqAck(remote *RemoteClient, size uint64, hash []byte) {
    if remote.download_state == HALF_OPEN {
        remote.download = InitDownload(remote.addr.IP.String(), remote.addr.Port, remote.download_rid, size, hash)
        remote.download_state = OPEN

        this.DownloadNextBlock(remote)
    }
}

func (this *Peer) HandleGet(remote *RemoteClient, index uint32) {
    if remote.upload_state != OPEN {
        this.conn.WriteToUDP(EncodeRst(), remote.addr)
        return
    }

    if remote.upload_rid == "file-list" {
        _, hash := GetFileListSizeAndHash(GetFileDir())
        if !bytes.Equal(remote.upload_original_hash, hash) {
            this.conn.WriteToUDP(EncodeRst(), remote.addr);
            log.Println("sent RST")
            return;
        }

        data_block, err := getFileListDataBlock(GetFileDir(), index)
        if err != nil {
            fmt.Println(err)
            return
        }

        this.conn.WriteToUDP(EncodeData(index, data_block), remote.addr)

    } else {
        // file download
        hashCheck := GetFileHash(remote.upload_file_path)

        // Checks if the file has been modified.
        if bytes.Equal(remote.upload_original_hash, hashCheck) {

            data_block, err := GetDataBlock(remote.upload_rid, index)
            if err == nil {
                this.conn.WriteToUDP(EncodeData(index, data_block), remote.addr)
            } else {
                fmt.Println(err.Error())
            }

            if (uint64(index + 1) * DATA_BLOCK_SIZE >= remote.upload_file_size) {
                remote.upload_state = CLOSED
            }
        } else {
            this.conn.WriteToUDP(EncodeRst(), remote.addr)
            log.Println("sent RST")
            return
        }
    }

}

func (this *Peer) HandleData(remote *RemoteClient, index uint32, data []byte) {
    if remote.download_state != OPEN {
        return
    }

    if remote.download.IndexPercent(index) < remote.download.IndexPercent(index + 1) {
        fmt.Println(remote.download.IndexPercent(index + 1), "%")
    }

    remote.download_deadline = time.Now().Add(time.Second * DEADLINE_SECONDS)
    remote.download.SaveData(index, data)
    this.DownloadNextBlock(remote)
}

func (this *Peer) HandlePacket(sender_addr *net.UDPAddr, packet_buffer []byte, packet_size int) {
    if this.remote_filter != nil && this.remote_filter.String() != sender_addr.String() {
        return
    }

    if !VerifyPacket(packet_buffer, packet_size) {
        return
    }

    packet_type := GetPacketType(packet_buffer)
    remote := this.GetRemote(sender_addr)
    switch packet_type {
    case REQ:
        log.Println("received REQ")
        err, rid := DecodeReq(packet_buffer, packet_size)
        if err == nil {
            this.HandleReq(remote, rid)
        }
    case REQ_ACK:
        log.Println("received REQ_ACK")
        err, size, hash := DecodeReqAck(packet_buffer, packet_size)
        if err == nil {
            this.HandleReqAck(remote, size, hash)
        }
    case REQ_NACK:
        log.Println("received REQ_NACK")
        log.Println("File does not exist or is not currently available.")
        remote.download_state = CLOSED

    case PUSH:
        log.Println("received PUSH")
        err, rid := DecodePush(packet_buffer, packet_size)
        if err == nil {
            this.HandlePush(remote, rid)
        }
    case PUSH_ACK:
        log.Println("received PUSH_ACK")
        remote.push_rid = ""

    case GET:
        err, index := DecodeGet(packet_buffer, packet_size)
        if err == nil {
            this.HandleGet(remote, index)
        }
    case DATA:
        err, index, data := DecodeData(packet_buffer, packet_size)
        if err == nil {
            this.HandleData(remote, index, data)
        }
    case RST:
        log.Println("received RST")
        log.Println("connection was reset, restarting download")

        // Resending REQ to restart download.
        remote.download_state = HALF_OPEN
        this.conn.WriteToUDP(EncodeReq(remote.download_rid), sender_addr)
        log.Println("sent REQ")
    case CNTF:
        err, info_byte, chunk_rid := DecodeCntf(packet_buffer, packet_size)
        if err == nil {
            log.Println("received CNTF - ACK " + chunk_rid)

            // end communication after CNTF was returned
            if (remote.cntf_state == OPEN) {
                // packet reached the node
                remote.cntf_state = CLOSED
            } else {
                // packet reached tornet
                id, file_path, sender_node := parseChunkRID(chunk_rid)
                chunk := this.torrent_map[file_path].ChunksMap[strconv.Itoa(id)]

                if (info_byte == 1) {
                    // add corresponding node to chunk holders
                    chunk.Nodes = append(chunk.Nodes, sender_node)
                    this.torrent_map[file_path].ChunksMap[strconv.Itoa(id)] = chunk
                    log.Println("added node from chunk. ", chunk)
                } else if (info_byte == 0) {
                    // remove corresponding node from chunk holders
                    node_index := -1
                    for i, node := range chunk.Nodes {
                        if node == sender_node {
                            node_index = i
                            break
                        }
                    }
                    if node_index != -1 {
                        chunk.Nodes = append(chunk.Nodes[:node_index], chunk.Nodes[node_index + 1:]...)
                        this.torrent_map[file_path].ChunksMap[strconv.Itoa(id)] = chunk
                    }
                    log.Println("deleted node from chunk. ", chunk)
                }
                this.conn.WriteToUDP(EncodeCntf(chunk_rid, info_byte), remote.addr)
            }
        }
    default:
        log.Println("dropping packet with invalid type", packet_type)
    }
}

func parseChunkRID(chunk_rid string) (int, string, string) {
    path_parts := strings.Split(chunk_rid, ".")
    index_str := path_parts[len(path_parts) - 1][len("part"):]
    index, _ := strconv.Atoi(index_str)

    file_path := strings.Join(path_parts[:len(path_parts) - 1], ".")
    _, file_name := filepath.Split(file_path)
    node_addr := strings.SplitN(file_path, "/", 2)
    return index, file_name, node_addr[0]
}

func (this *Peer) CheckTimeouts() {
    now := time.Now()

    for _, client := range this.remotes {

        if client.download_deadline.Before(now) && client.download_state != CLOSED {
            if client.download_state == HALF_OPEN {
                this.conn.WriteToUDP(EncodeReq(client.download_rid), client.addr)
                log.Println("sent REQ")
            } else {
                // state == OPEN
                client.download.ResetGet()
                this.conn.WriteToUDP(client.download.CreateNextGet(), client.addr)
                log.Println("sent GET")
            }

            client.download_deadline = time.Now().Add(time.Second * DEADLINE_SECONDS)
        }

        if client.push_rid != "" && client.push_deadline.Before(now) {
            this.conn.WriteToUDP(EncodePush(client.push_rid), client.addr)
            client.push_deadline = time.Now().Add(time.Second * DEADLINE_SECONDS)
        }
        if client.cntf_deadline.Before(now) && client.cntf_state != CLOSED {
            this.conn.WriteToUDP(EncodeCntf(client.cntf_rid, 0), client.addr)
            log.Println("sent CNTF")
        }
    }
}

func (this *Peer) ReadLoop() {
    buf := make([]byte, UDP_BUFFER_SIZE)

    for {
        packet_size, sender_addr, err := this.conn.ReadFromUDP(buf)
        if err == nil {
            this.read_chan <- InboundPacket{sender_addr, buf, packet_size}
        }
    }
}

func (this *Peer) Run() {
    go this.ReadLoop()

    regular_ticker := time.NewTicker(time.Second * 1).C

    // this is flexible: we can add a timer for congestion control
    // do lock and wait for now: send next get when data packet arrived

    for {
        select {
        case <-regular_ticker:
            this.CheckTimeouts()
        case packet := <-this.read_chan:
            this.HandlePacket(packet.sender, packet.data, packet.size)

            remote := this.GetRemote(packet.sender)
            if remote.download_state == CLOSED && remote.upload_state == CLOSED {
                return
            }
        }
    }
}

func (this *Peer) Download(rid string, remote_addr *net.UDPAddr) {
    fmt.Println("downloading", rid)

    remote := this.GetRemote(remote_addr)

    remote.download_rid = rid
    remote.download_state = HALF_OPEN
    remote.download_deadline = time.Now().Add(time.Second * DEADLINE_SECONDS)
    this.conn.WriteToUDP(EncodeReq(rid), remote_addr)
    log.Println("sent REQ")
}

func (this *Peer) Upload(rid string, remote_addr *net.UDPAddr) {
    remote := this.GetRemote(remote_addr)
    remote.push_deadline = time.Now().Add(time.Second * DEADLINE_SECONDS)
    remote.push_rid = rid
    remote.upload_state = OPEN

    this.conn.WriteToUDP(EncodePush(rid), remote.addr)
    log.Println("sending push for", rid)
}

func (this *Peer) SendChunkNotification(rid string, info_byte byte, remote_addr *net.UDPAddr) {
    remote := this.GetRemote(remote_addr)
    remote.cntf_deadline = time.Now().Add(time.Second * DEADLINE_SECONDS)
    remote.cntf_rid = rid
    remote.cntf_state = OPEN
    this.conn.WriteToUDP(EncodeCntf(rid, info_byte), remote.addr)
    log.Println("sending push for", rid)
}

func printFileList() {
    path := filepath.Join(GetFileDir(), "file-list")
    f, err := os.Open(path)
    if err != nil {
        fmt.Println(err.Error())
    }
    defer f.Close()

    var lines []string
    scanner := bufio.NewScanner(f)
    for scanner.Scan() {
        lines = append(lines, scanner.Text())
    }

    fmt.Println("Files on server:")
    for _, f := range lines {
        fmt.Println("\t" + f)
    }
}

