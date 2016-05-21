package pft

import (
    "net"
    "fmt"
    "time"
    "os"
    "log"
    "strings"
    "bufio"
)

type InboundPacket struct {
    sender *net.UDPAddr
    data   []byte
    size   int
}

type RemoteClient struct {
    download_deadline time.Time // if this deadline is passed, resend last packet
    download_state    int
    download_rid      string    // for when state == HALF_OPEN: REQ or it's answer may get lost, so we need this for timeouts
    download          *Download  // this is for ACK'ed downloads, do not put download_state and download_rid into Download struct

    upload_state      int
    upload_rid        string
}

type Peer struct {
    remote_filter *net.UDPAddr // if remote_filter != nil: only accept packets from this address (client mode)
    conn          *net.UDPConn
    read_chan     chan InboundPacket
    remotes       map[string]*RemoteClient
}

func MakePeer(localAddr *net.UDPAddr, remoteAddr *net.UDPAddr) Peer {
    conn, _ := net.ListenUDP("udp", localAddr)

    return Peer{
        remote_filter: remoteAddr,
        conn: conn,
        read_chan: make(chan InboundPacket),
        remotes: make(map[string]*RemoteClient),
    }
}

func (this Peer) HandleReq(sender *net.UDPAddr, rid string) {
    remote := new(RemoteClient)
    this.remotes[sender.String()] = remote
    remote.upload_rid = rid
    remote.upload_state = CLOSED

    if strings.HasPrefix(rid, "file:") {
        upload_file_path := fmt.Sprintf("%s/%s", GetFileDir(), rid[5:len(rid)])

        f, err := os.Open(upload_file_path)
        stat, err := f.Stat()
        if err != nil {
            log.Println("Error: " + err.Error())
            this.conn.WriteToUDP(EncodeReqNack(), sender)
            log.Println("sent REQ-NACK")
            return
        }

        hash := GetFileHash(upload_file_path)
        req_ack := EncodeReqAck(uint64(stat.Size()), hash)
        this.conn.WriteToUDP(req_ack, sender)
        remote.upload_state = OPEN
        log.Println("sent REQ-ACK")
    } else if rid == "file-list" {
        size, hash := GetFileListSizeAndHash(GetFileDir())
        req_ack := EncodeReqAck(size, hash)
        this.conn.WriteToUDP(req_ack, sender)
        remote.upload_state = OPEN
        log.Println("sent REQ-ACK")
    } else {
        this.conn.WriteToUDP(EncodeReqNack(), sender)
    }
}

func (this Peer) HandleReqAck(sender *net.UDPAddr, size uint64, hash []byte) {
    remote := this.remotes[sender.String()]
    if remote.download_state == HALF_OPEN {
        remote.download = InitDownload(sender.IP.String(), sender.Port, remote.download_rid, size, hash)
        this.conn.WriteToUDP(remote.download.CreateNextGet(), sender)
        remote.download_state = OPEN
    }
}

func (this Peer) HandleGet(sender *net.UDPAddr, index uint32) {
    remote := this.remotes[sender.String()]
    if remote.upload_state == OPEN {
        data_block, err := GetDataBlock(remote.upload_rid, index)
        if err == nil {
            this.conn.WriteToUDP(EncodeData(index, data_block), sender)
        } else {
            fmt.Println(err.Error())
        }
    }
}

func (this Peer) HandleData(sender *net.UDPAddr, index uint32, data []byte) {
    remote := this.remotes[sender.String()]
    if remote.download_state == OPEN {
        remote.download.HandleDataPacket(index, data)
        if remote.download.IsFinished() {
            remote.download.Close()
            if (remote.download_rid == "file-list") {
                printFileList()
            }
            os.Exit(0)
        } else {
            this.conn.WriteToUDP(remote.download.CreateNextGet(), sender)
        }
    }
}

func (this Peer) HandlePacket(sender_addr *net.UDPAddr, packet_buffer []byte, packet_size int) {
    if this.remote_filter != nil && this.remote_filter.String() != sender_addr.String() {
        return
    }

    if !VerifyPacket(packet_buffer, packet_size) {
        return
    }

    packet_type := GetPacketType(packet_buffer)
    switch packet_type {
    case REQ:
        log.Println("received REQ")
        err, rid := DecodeReq(packet_buffer, packet_size)
        if err == nil {
            this.HandleReq(sender_addr, rid)
        }
    case REQ_ACK:
        log.Println("received REQ_ACK")
        err, size, hash := DecodeReqAck(packet_buffer, packet_size)
        if err == nil {
            this.HandleReqAck(sender_addr, size, hash)
        }
    case REQ_NACK:
        log.Println("received REQ_NACK")
    case PUSH:
        // TODO
        log.Println("received PUSH")
    case PUSH_ACK:
        // TODO
        log.Println("received PUSH_ACK")
    case GET:
        log.Println("received GET")
        err, index := DecodeGet(packet_buffer, packet_size)
        if err == nil {
            this.HandleGet(sender_addr, index)
        }
    case DATA:
        log.Println("received DATA")
        time.Sleep(time.Second * 2)
        err, index, data := DecodeData(packet_buffer, packet_size)
        if err == nil {
            this.HandleData(sender_addr, index, data)
        }
    case RST:
        // TODO
        log.Println("received RST")
    default:
        log.Println("dropping packet with invalid type", packet_type)
    }
}

func (this Peer) CheckTimeouts() {
    now := time.Now()

    for addrString, _ := range this.remotes {
        client := this.remotes[addrString] // get by reference so we can update it, range would only get it by value
        address, _ := net.ResolveUDPAddr("udp", addrString)

        if client.download_deadline.Before(now) && client.download_state != CLOSED {
            if client.download_state == HALF_OPEN {
                this.conn.WriteToUDP(EncodeReq(client.download_rid), address)
                log.Println("sent REQ")
            } else {
                // state == OPEN
                client.download.ResetGet()
                this.conn.WriteToUDP(client.download.CreateNextGet(), address)
                log.Println("sent GET")
            }

            client.download_deadline = time.Now().Add(time.Second * 4)
        }
    }
}

func (this Peer) ReadLoop() {
    buf := make([]byte, UDP_BUFFER_SIZE)

    for {
        packet_size, sender_addr, err := this.conn.ReadFromUDP(buf)
        if err == nil {
            this.read_chan <- InboundPacket{sender_addr, buf, packet_size}
        }
    }
}

func (this Peer) Run() {
    go this.ReadLoop()

    check_timeouts := time.NewTicker(time.Second * 1).C

    // this is flexible: we can add a timer for congestion control
    // do lock and wait for now: send next get when data packet arrived

    for {
        select {
        case <-check_timeouts:
            this.CheckTimeouts()
        case packet := <-this.read_chan:
            this.HandlePacket(packet.sender, packet.data, packet.size)
        }
    }
}

func (this Peer) Download(rid string, remote_addr *net.UDPAddr) {
    remote := new(RemoteClient)
    this.remotes[remote_addr.String()] = remote

    remote.download_rid = rid
    remote.download_state = HALF_OPEN
    remote.download_deadline = time.Now().Add(time.Second * 4)
    this.conn.WriteToUDP(EncodeReq(rid), remote_addr)
    log.Println("sent REQ")
}

func printFileList() {
    path := fmt.Sprintf("%s/file-list", GetFileDir())
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
