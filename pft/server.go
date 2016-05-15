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


func Server(port int) {
	addr, err := net.ResolveUDPAddr("udp", ":" + strconv.Itoa(port))
	CheckError(err)

	conn, err := net.ListenUDP("udp", addr)
	CheckError(err)
	fmt.Println("listening on", addr)
	defer conn.Close()

	storage_dir := "./server_files"
	file_list_mode := false
	local_file_path := ""
	current_state := CLOSED
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

			var req_ack []byte
			if strings.HasPrefix(rid, "file:") {
				filename := rid[5:len(rid)]
				local_file_path = fmt.Sprintf("%s/%s", storage_dir, filename)

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
				req_ack = EncodeReqAck(uint64(stat.Size()), hash)
			} else if rid == "file-list" {
				file_list_mode = true
				size, hash := GetFileListSizeAndHash(storage_dir)
				req_ack = EncodeReqAck(size, hash)
			} else {
				CheckError(errors.New("unknown resource type"))
			}

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

			var data_block []byte
			if file_list_mode {
				data_block = GetFileListDataBlock(storage_dir, index)
			} else {
				data_block = GetFileDataBlock(local_file_path, index)
			}

			data := EncodeData(index, data_block)
			conn.WriteToUDP(data, sender_addr)
			log.Println("Sent DATA: ", data)
		}
	}
}