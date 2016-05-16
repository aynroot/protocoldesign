package pft

import (
	"errors"
	"fmt"
	"log"
	"net"
	"strconv"
	"time"
	"os"
)

type client struct {
	server_addr *net.UDPAddr
	storage_dir string
	state       int
	conn        *net.UDPConn
	download    *Download
	resource    string
	get_next    bool
}

func initClient(conn *net.UDPConn, server_addr *net.UDPAddr, resource string) *client {
	return &client{
		server_addr: server_addr,
		storage_dir: "./client_files",
		state:       CLOSED,
		conn:        conn,
		download:    nil,
		resource:    resource,
		get_next:    true,
	}
}

func (this *client) checkNetError(err error) error {
	if neterr, ok := err.(net.Error); ok && neterr.Timeout() {
		log.Println(neterr)

		this.download.SavePartialDownloadInfo()
		this.state = CLOSED
		this.conn.Close()
		return neterr
	} else {
		CheckError(err)
	}
	return nil
}

func (this *client) sendReq(server string, port int) {
	exists, info_file_path := CheckIfPartiallyDownloaded(server, port, this.resource)
	if exists {
		this.download = LoadPartFile(info_file_path)
	} else {
		this.download = InitDownload(server, port, this.resource, this.storage_dir)
	}
	log.Println(this.download)

	req := EncodeReq(this.resource)
	this.conn.WriteToUDP(req, this.server_addr)
	log.Println("Sent REQ:", req)
	this.state = HALF_OPEN
}

func (this *client) handleReqResponse(packet []byte, packet_size int) error {
	if !VerifyPacket(packet, packet_size) {
		log.Println("Verification (REQ_ACK) failed")
		return nil
	}
	packet_type := GetPacketType(packet)
	if packet_type == REQ_ACK {
		err, size, hash := DecodeReqAck(packet, packet_size)
		CheckError(err)

		this.download.HandleReqPacket(uint64(size), hash)
		this.state = OPEN
	} else if packet_type == REQ_NACK {
		this.state = CLOSED
	} else {
		CheckError(errors.New("undeexpected packet type"))
	}
	return nil
}

func (this *client) receiveData() error {
	if this.get_next {
		get := this.download.CreateNextGet()
		this.conn.WriteToUDP(get, this.server_addr)
		log.Println("Sent GET:", get)
	}

	buf := make([]byte, UDP_BUFFER_SIZE)
	packet_size, _, err := this.conn.ReadFromUDP(buf)
	if err != nil {
		return this.checkNetError(err)
	}

	if !VerifyPacket(buf, packet_size) {
		log.Println("Verification (DATA) failed")
		return nil
	}
	err, index, data := DecodeData(buf, packet_size)
	CheckError(err)

	this.get_next = true
	if !this.download.HandleDataPacket(index, data) {
		log.Println("Data is not written on disk")
		this.get_next = false
	}
	if this.download.IsFinished() {
		this.download.FinishDownload()
		if this.resource == "file-list" {
			file_list, err := ReturnFileList()
			CheckError(err)
			fmt.Println("Files on server:")
			for _, f := range file_list {
				fmt.Println("\t" + f)
			}
		} else {
			fmt.Printf("%s is successfully downloaded.\n", this.resource)
		}
		os.Exit(0)
	}
	return nil
}

func Client(port int, server string, resource string) {
	server_addr, err := net.ResolveUDPAddr("udp", server+":"+strconv.Itoa(port))
	CheckError(err)

	local_addr, err := net.ResolveUDPAddr("udp", "127.0.0.1:0")
	CheckError(err)

	conn, err := net.ListenUDP("udp", local_addr)
	CheckError(err)
	defer conn.Close()

	conn.SetDeadline(time.Now().Add(time.Second * 10))

	client := *initClient(conn, server_addr, resource)
	for {
		if client.state == CLOSED {
			client.sendReq(server, port)
		}
		if client.state == HALF_OPEN {
			buf := make([]byte, UDP_BUFFER_SIZE)
			packet_size, _, err := client.conn.ReadFromUDP(buf)
			if err != nil {
				if client.checkNetError(err) != nil {
					fmt.Println("Network timeout")
					os.Exit(0)
				}
			}
			client.handleReqResponse(buf, packet_size)
		} else if client.state == OPEN {
			err = client.receiveData()
			if err != nil {
				fmt.Println("Network timeout")
				os.Exit(0)
			}
		}
	}
}
