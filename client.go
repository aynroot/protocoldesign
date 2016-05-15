package main

import (
	"log"
	"fmt"
	"os"
	"net"
	"strconv"
	"github.com/aynroot/protocoldesign/pft"
)

func Client(port int, server string, resource string) {
	server_addr, err := net.ResolveUDPAddr("udp", server + ":" + strconv.Itoa(port))
	CheckError(err)

	local_addr, err := net.ResolveUDPAddr("udp", "127.0.0.1:0")
	CheckError(err)

	conn, err := net.ListenUDP("udp", local_addr)
	CheckError(err)
	defer conn.Close()

	current_state := pft.CLOSED

	get_next := true
	var download *pft.Download
	for {
		if current_state == pft.CLOSED {
			storage_dir := "./client_files"
			exists, info_file_path := pft.CheckIfPartiallyDownloaded(server, port, resource)
			download = new(pft.Download)
			if exists {
				download = pft.LoadPartialDownload(info_file_path)
			} else {
				download = pft.InitDownload(server, port, resource, storage_dir)
			}
			log.Println(download)

			req := pft.EncodeReq(resource)
			conn.WriteToUDP(req, server_addr)
			log.Println("Sent REQ:", req)
			current_state = pft.HALF_OPEN
		}
		if current_state == pft.HALF_OPEN {
			buf := make([]byte, UDP_BUFFER_SIZE)
			packet_size, _, err := conn.ReadFromUDP(buf)
			CheckError(err)

			if !pft.VerifyPacket(buf, packet_size) {
				log.Println("Verification (REQ_ACK) failed")
				continue
			}
			packet_type := pft.GetPacketType(buf)
			if packet_type == pft.REQ {
				err, size, hash := pft.DecodeReqAck(buf, packet_size)
				CheckError(err)

				download.HandleReqPacket(uint64(size), hash)
				current_state = pft.OPEN
			} else if packet_type == pft.NACK {
				current_state = pft.CLOSED
			} else {
				fmt.Println("Error: undeexpected packet type")
				os.Exit(0)
			}

		} else if current_state == pft.OPEN {
			if get_next {
				get := download.CreateNextGet()
				conn.WriteToUDP(get, server_addr)
				log.Println("Sent GET:", get)
			}

			buf := make([]byte, UDP_BUFFER_SIZE)
			packet_size, _, err := conn.ReadFromUDP(buf)
			CheckError(err)

			if !pft.VerifyPacket(buf, packet_size) {
				log.Println("Verification (DATA) failed")
				continue
			}
			err, index, data := pft.DecodeData(buf, packet_size)
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
