package pft

import (
	"log"
	"strings"
	"fmt"
	"os"
	"net"
	"strconv"
)


func Server(port int) {
	addr, err := net.ResolveUDPAddr("udp", ":" + strconv.Itoa(port))
	CheckError(err)

	conn, err := net.ListenUDP("udp", addr)
	CheckError(err)
	fmt.Println("listening on", addr)
	defer conn.Close()

	current_state := CLOSED
	var local_file_path string
	for {
		if current_state == CLOSED {
			buf := make([]byte, UDP_BUFFER_SIZE)
			packet_size, sender_addr, err := conn.ReadFromUDP(buf)
			CheckError(err)

			if !VerifyPacket(buf, packet_size) {
				log.Println("Verification (REQ) failed")
				continue
			}
			_, rid := DecodeReq(buf, packet_size)
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
				nack := EncodeNack()
				conn.WriteToUDP(nack, sender_addr)
				log.Println("Sent NACK:", nack)
				continue
			}

			hash := GetFileHash(local_file_path)
			req_ack := EncodeReqAck(uint64(stat.Size()), hash)

			conn.WriteToUDP(req_ack, sender_addr)
			log.Println("Sent REQ-ACK:", req_ack)
			current_state = OPEN
		} else if current_state == OPEN {
			buf := make([]byte, UDP_BUFFER_SIZE)
			packet_size, sender_addr, err := conn.ReadFromUDP(buf)
			CheckError(err)

			if !VerifyPacket(buf, packet_size) {
				log.Println("Verification (GET) failed")
				continue
			}
			err, index := DecodeGet(buf, packet_size)
			CheckError(err)

			log.Println("Received GET:", index)
			data_block := GetFileDataBlock(local_file_path, index)
			data := EncodeData(index, data_block)
			conn.WriteToUDP(data, sender_addr)
			log.Println("Sent DATA: ", data)
		}
	}
}