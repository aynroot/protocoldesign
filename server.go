package main

import (
	"log"
	"strings"
	"fmt"
	"os"
	"net"
	"strconv"
	"github.com/aynroot/protocoldesign/pft"
)


func Server(port int) {
	addr, err := net.ResolveUDPAddr("udp", ":" + strconv.Itoa(port))
	CheckError(err)

	conn, err := net.ListenUDP("udp", addr)
	CheckError(err)
	fmt.Println("listening on", addr)
	defer conn.Close()

	current_state := pft.CLOSED
	var local_file_path string
	for {
		if current_state == pft.CLOSED {
			buf := make([]byte, UDP_BUFFER_SIZE)
			packet_size, sender_addr, err := conn.ReadFromUDP(buf)
			CheckError(err)

			if !pft.VerifyPacket(buf, packet_size) {
				log.Println("Verification (REQ) failed")
				continue
			}
			_, rid := pft.DecodeReq(buf, packet_size)
			log.Println("Received REQ:", rid)

			// file or file-list // TODO: handle file-list
			storage_dir := "./server_files"
			if strings.HasPrefix(rid, "file:") {
				filename := rid[5:len(rid)]
				local_file_path = fmt.Sprintf("%s/%s", storage_dir, filename)
			}

			f, err := os.Open(local_file_path)
			stat, err := f.Stat()
			if err != nil {
				log.Println("Error: " + err.Error())
				nack := pft.EncodeNack()
				conn.WriteToUDP(nack, sender_addr)
				log.Println("Sent NACK:", nack)
				continue
			}

			hash := pft.GetFileHash(local_file_path)
			req_ack := pft.EncodeReqAck(uint64(stat.Size()), hash)

			conn.WriteToUDP(req_ack, sender_addr)
			log.Println("Sent REQ-ACK:", req_ack)
			current_state = pft.OPEN
		} else if current_state == pft.OPEN {
			buf := make([]byte, UDP_BUFFER_SIZE)
			packet_size, sender_addr, err := conn.ReadFromUDP(buf)
			CheckError(err)

			if !pft.VerifyPacket(buf, packet_size) {
				log.Println("Verification (GET) failed")
				continue
			}
			err, index := pft.DecodeGet(buf, packet_size)
			CheckError(err)

			log.Println("Received GET:", index)
			data_block := pft.GetFileDataBlock(local_file_path, index)
			data := pft.EncodeData(index, data_block)
			conn.WriteToUDP(data, sender_addr)
			log.Println("Sent DATA: ", data)
		}
	}
}