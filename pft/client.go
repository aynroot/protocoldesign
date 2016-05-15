package pft

import (
	"log"
	"fmt"
	"os"
	"net"
	"strconv"
)

const UDP_BUFFER_SIZE = 512

func CheckError(err error) {
	if err != nil {
		fmt.Println("Error: ", err)
		os.Exit(0)
	}
}

func Client(port int, server string, resource string) {
	server_addr, err := net.ResolveUDPAddr("udp", server + ":" + strconv.Itoa(port))
	CheckError(err)

	local_addr, err := net.ResolveUDPAddr("udp", "127.0.0.1:0")
	CheckError(err)

	conn, err := net.ListenUDP("udp", local_addr)
	CheckError(err)
	defer conn.Close()

	current_state := CLOSED

	get_next := true
	var download *Download
	for {
		if current_state == CLOSED {
			storage_dir := "./client_files"
			exists, info_file_path := CheckIfPartiallyDownloaded(server, port, resource)
			download = new(Download)
			if exists {
				download = LoadPartialDownload(info_file_path)
			} else {
				download = InitDownload(server, port, resource, storage_dir)
			}
			log.Println(download)

			req := EncodeReq(resource)
			conn.WriteToUDP(req, server_addr)
			log.Println("Sent REQ:", req)
			current_state = HALF_OPEN
		}
		if current_state == HALF_OPEN {
			buf := make([]byte, UDP_BUFFER_SIZE)
			packet_size, _, err := conn.ReadFromUDP(buf)
			CheckError(err)

			if !VerifyPacket(buf, packet_size) {
				log.Println("Verification (REQ_ACK) failed")
				continue
			}
			packet_type := GetPacketType(buf)
			if packet_type == REQ_ACK {
				err, size, hash := DecodeReqAck(buf, packet_size)
				CheckError(err)

				download.HandleReqPacket(uint64(size), hash)
				current_state = OPEN
			} else if packet_type == NACK {
				current_state = CLOSED
			} else {
				fmt.Println("Error: undeexpected packet type")
				os.Exit(0)
			}

		} else if current_state == OPEN {
			if get_next {
				get := download.CreateNextGet()
				conn.WriteToUDP(get, server_addr)
				log.Println("Sent GET:", get)
			}

			buf := make([]byte, UDP_BUFFER_SIZE)
			packet_size, _, err := conn.ReadFromUDP(buf)
			CheckError(err)

			if !VerifyPacket(buf, packet_size) {
				log.Println("Verification (DATA) failed")
				continue
			}
			err, index, data := DecodeData(buf, packet_size)
			CheckError(err)

			get_next = true
			if (!download.HandleDataPacket(index, data)) {
				log.Println("Data is not written on disk")
				get_next = false
			}
			if download.IsFinished() {
				download.FinishDownload()
				fmt.Println("Download is successfully finished.")
				break
			}
		}
	}
}
