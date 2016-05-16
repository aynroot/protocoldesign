package pft

import (
    "net"
    "fmt"
)


type Peer struct {
    download_state int
    upload_state   int
    file_dir       string
    remote_filter  *net.UDPAddr  // only accept packets from this address (client mode)
    conn           *net.UDPConn
    download_queue chan Download
}

func MakePeer(localAddr *net.UDPAddr, remoteAddr *net.UDPAddr) Peer {
    conn, _ := net.ListenUDP("udp", localAddr)

    return Peer {
        download_state: CLOSED,
        upload_state: CLOSED,
        file_dir: ".",
        remote_filter: remoteAddr,
        conn: conn,
        download_queue: make(chan Download, 10)}
}


func (this Peer) HandleReq(sender *net.UDPAddr, req []byte, size int) {
    err, rid := DecodeReq(req, size)
    if err != nil {
        fmt.Println("Error decoding req:", err)
        return
    }



}


func (this Peer) ReadLoop() {
    buf := make([]byte, UDP_BUFFER_SIZE)

    for {
        packet_size, sender_addr, _ := server.conn.ReadFromUDP(buf)
        if this.remote_filter != nil && this.remote_filter != sender_addr {
            continue
        }

        if !VerifyPacket(buf, packet_size) {
            continue
        }

        packet_type := buf[16]
        switch packet_type {
        case REQ:
        case REQ_ACK:
        case REQ_NACK:
        case PUSH:
        case PUSH_ACK:
        case GET:
        case DATA:
        case RST:
        default:
            fmt.Println("dropping packet with invalid type", packet_type)
            continue
        }
    }
}


func (this Peer) Download(rid string, remote *net.UDPAddr) {
}

func (this Peer) WriteLoop() {
    for {
        switch this.download_state {
        case CLOSED:

        }
    }
}

func (this Peer) Run() {
    go this.WriteLoop()
    go this.ReadLoop()
}



