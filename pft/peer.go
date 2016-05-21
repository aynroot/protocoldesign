package pft

import (
    "net"
    "fmt"
    "time"
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
    download          Download  // this is for ACK'ed downloads, do not put download_state and download_rid into Download struct

    upload_state      int
    upload_rid        string
}

type Peer struct {
    remote_filter *net.UDPAddr // if remote_filter != nil: only accept packets from this address (client mode)
    conn          *net.UDPConn
    read_chan     chan InboundPacket
    remotes       map[*net.UDPAddr]*RemoteClient
}

func MakePeer(localAddr *net.UDPAddr, remoteAddr *net.UDPAddr) Peer {
    conn, _ := net.ListenUDP("udp", localAddr)

    return Peer{
        remote_filter: remoteAddr,
        conn: conn,
        read_chan: make(chan InboundPacket),
        remotes: make(map[*net.UDPAddr]*RemoteClient),
    }
}

func (this Peer) HandleReq(sender *net.UDPAddr, rid string) {
    // todo, handle req and put state into correct RemoteClient, then answer req

    this.conn.WriteToUDP(EncodeReqNack(), sender)
}

func (this Peer) HandlePacket(sender_addr *net.UDPAddr, packet_buffer []byte, packet_size int) {
    if this.remote_filter != nil && this.remote_filter != sender_addr {
        return
    }

    if !VerifyPacket(packet_buffer, packet_size) {
        return
    }

    sender_client := this.remotes[sender_addr]

    packet_type := packet_buffer[16]
    switch packet_type {
    case REQ:
        fmt.Println("received REQ")
        err, rid := DecodeReq(packet_buffer, packet_size)
        if err == nil {
            this.HandleReq(sender_addr, rid)
        }
    case REQ_ACK:
        fmt.Println("received REQ_ACK")
        err, size, hash := DecodeReqAck(packet_buffer, packet_size)

        if err == nil && sender_client.download_state == HALF_OPEN {
            sender_client.download = InitDownload(sender_addr.String(), sender_addr.Port,
                sender_client.download_rid, size, hash)
            this.conn.WriteToUDP(this.remotes[sender_addr].download.CreateNextGet(), sender_addr)
        }
    case REQ_NACK:
        fmt.Println("received REQ_NACK")
    // todo
    case PUSH:
        fmt.Println("received PUSH")
    // todo
    case PUSH_ACK:
        fmt.Println("received PUSH_ACK")
    // todo
    case GET:
        fmt.Println("received GET")
    // todo
    case DATA:
        fmt.Println("received DATA")
    // todo
    case RST:
        fmt.Println("received RST")
    // todo
    default:
        fmt.Println("dropping packet with invalid type", packet_type)
    }
}

func (this Peer) CheckTimeouts() {
    now := time.Now()

    for address, _ := range this.remotes {
        client := this.remotes[address] // get by reference so we can update it, range would only get it by value

        if client.download_deadline.Before(now) && client.download_state != CLOSED {
            if client.download_state == HALF_OPEN {
                this.conn.WriteToUDP(EncodeReq(client.download_rid), address)
            } else {
                // state == OPEN
                client.download.ResetGet()
                this.conn.WriteToUDP(client.download.CreateNextGet(), address)
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
    this.remotes[remote_addr] = remote

    remote.download_rid = rid
    remote.download_state = HALF_OPEN
    remote.download_deadline = time.Now().Add(time.Second * 4)
    this.conn.WriteToUDP(EncodeReq(rid), remote_addr)
}
