package pft

import (
	"log"
	"strings"
	"fmt"
	"os"
	"net"
	"strconv"
	"errors"
)

type server struct {
	storage_dir string
	local_file_path string
	file_list_mode bool
	state int
	conn *net.UDPConn
}

func initServer(conn *net.UDPConn) *server {
	return &server{
		storage_dir: "./server_files",
		local_file_path: "",
		file_list_mode: false,
		state: CLOSED,
		conn: conn,
	}
}

func (this *server) handleReq(packet []byte, packet_size int, sender_addr *net.UDPAddr) error {
	if !VerifyPacket(packet, packet_size) {
		log.Println("Verification (REQ) failed")
		return nil
	}
	_, rid := DecodeReq(packet, packet_size)
	log.Println("Received REQ:", rid)

	var req_ack []byte
	if strings.HasPrefix(rid, "file:") {
		filename := rid[5:len(rid)]
		this.local_file_path = fmt.Sprintf("%s/%s", this.storage_dir, filename)

		f, err := os.Open(this.local_file_path)
		stat, err := f.Stat()
		if err != nil {
			log.Println("Error: " + err.Error())
			nack := EncodeNack()
			this.conn.WriteToUDP(nack, sender_addr)
			log.Println("Sent NACK:", nack)
			return nil
		}

		hash := GetFileHash(this.local_file_path)
		req_ack = EncodeReqAck(uint64(stat.Size()), hash)
	} else if rid == "file-list" {
		this.file_list_mode = true
		size, hash := GetFileListSizeAndHash(this.storage_dir)
		req_ack = EncodeReqAck(size, hash)
	} else {
		CheckError(errors.New("unknown resource type"))
	}

	this.conn.WriteToUDP(req_ack, sender_addr)
	log.Println("Sent REQ-ACK:", req_ack)
	this.state = OPEN
	return nil
}

func (this *server) handleGet(packet []byte, packet_size int, sender_addr *net.UDPAddr) error {
	if !VerifyPacket(packet, packet_size) {
		log.Println("Verification (GET) failed")
		return nil
	}
	packet_type := GetPacketType(packet)
	if packet_type == GET {
		err, index := DecodeGet(packet, packet_size)
		CheckError(err)
		log.Println("Received GET:", index)

		var data_block []byte
		if this.file_list_mode {
			data_block = GetFileListDataBlock(this.storage_dir, index)
		} else {
			data_block = GetFileDataBlock(this.local_file_path, index)
		}

		data := EncodeData(index, data_block)
		this.conn.WriteToUDP(data, sender_addr)
		log.Println("Sent DATA: ", data)
	} else if packet_type == REQ {
		this.handleReq(packet, packet_size, sender_addr)
	} else {
		log.Printf("received invalid packet: %d\n", packet_type)
	}
	return nil
}

func Server(port int) {
	addr, err := net.ResolveUDPAddr("udp", ":" + strconv.Itoa(port))
	CheckError(err)

	conn, err := net.ListenUDP("udp", addr)
	CheckError(err)
	fmt.Println("listening on", addr)
	defer conn.Close()

	server := *initServer(conn)
	for {
		if server.state == CLOSED {
			buf := make([]byte, UDP_BUFFER_SIZE)
			packet_size, sender_addr, err := server.conn.ReadFromUDP(buf)
			CheckError(err)

			server.handleReq(buf, packet_size, sender_addr)
		} else if server.state == OPEN {
			buf := make([]byte, UDP_BUFFER_SIZE)
			packet_size, sender_addr, err := server.conn.ReadFromUDP(buf)
			CheckError(err)

			err = server.handleGet(buf, packet_size, sender_addr)
			if err != nil { // REQ received
				server = *initServer(conn)
			}
		}
	}
}