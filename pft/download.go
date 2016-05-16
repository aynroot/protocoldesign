package pft

import (
	"fmt"
	"log"
	"os"
	"strings"

	"bytes"
	"encoding/gob"
	"errors"
	"io/ioutil"
	"path/filepath"
	"bufio"
)

// this packet deals with resource downloads
// a file format for storing partial downloads needs to be developed
// a partial download file needs to save the rid, the size, the hash and the already downloaded data of the resource

type Download struct {
	server                string
	port                  int
	rid                   string
	size                  uint64
	hash                  []byte
	received_index        uint32
	requested_index       uint32
	partial_file_path     string
	destination_file_path string
}

func (this *Download) GobEncode() ([]byte, error) {
	buffer := new(bytes.Buffer)
	encoder := gob.NewEncoder(buffer)

	encoder.Encode(this.server)
	encoder.Encode(this.port)
	encoder.Encode(this.rid)
	encoder.Encode(this.size)
	encoder.Encode(this.hash)
	encoder.Encode(this.received_index)
	encoder.Encode(this.requested_index)
	encoder.Encode(this.partial_file_path)
	encoder.Encode(this.destination_file_path)

	return buffer.Bytes(), nil
}

func (this *Download) GobDecode(data []byte) error {
	buffer := bytes.NewBuffer(data)
	decoder := gob.NewDecoder(buffer)

	decoder.Decode(&this.server)
	decoder.Decode(&this.port)
	decoder.Decode(&this.rid)
	decoder.Decode(&this.size)
	decoder.Decode(&this.hash)
	decoder.Decode(&this.received_index)
	decoder.Decode(&this.requested_index)
	decoder.Decode(&this.partial_file_path)
	decoder.Decode(&this.destination_file_path)

	return nil
}

func ReturnFileList() ([]string, error) {
	path := fmt.Sprintf("./.pft/file-list")
	f, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer f.Close()

	var lines []string
	scanner := bufio.NewScanner(f)
	for scanner.Scan() {
		lines = append(lines, scanner.Text())
	}
	return lines, scanner.Err()
}

// returns true if the payload was written to disk
func (this *Download) HandleDataPacket(index uint32, payload []byte) bool {
	if index == this.received_index + 1 {
		file, err := os.OpenFile(this.partial_file_path, os.O_APPEND | os.O_WRONLY | os.O_CREATE, 0600)
		CheckError(err)

		n, err := file.Write(payload)
		CheckError(err)
		log.Printf("payload: wrote %d bytes\n", n)
		file.Close()

		this.received_index = index
		return true
	} else {
		this.requested_index = this.received_index
	}
	return false
}

func (this *Download) HandleReqPacket(size uint64, hash []byte) {
	this.size = size
	this.hash = hash
}

// saves the internal variables to a partial download file
func (this *Download) SavePartialDownloadInfo() {
	this.requested_index = this.received_index
	file, err := os.Create(this.partial_file_path + ".info")
	CheckError(err)
	defer file.Close()

	buffer := bytes.Buffer{}
	encoder := gob.NewEncoder(&buffer)
	err = encoder.Encode(this)
	CheckError(err)

	n, err := file.Write(buffer.Bytes())
	CheckError(err)
	log.Printf("info: wrote %d bytes\n", n)
}

// loads the internal variables from a partial download file
func LoadPartialDownload(info_file_path string) *Download {
	data, err := ioutil.ReadFile(info_file_path)
	CheckError(err)

	download_info := new(Download)
	buffer := bytes.NewBuffer(data)
	decoder := gob.NewDecoder(buffer)
	err = decoder.Decode(&download_info)
	CheckError(err)

	log.Println(download_info)
	return download_info
}

func (this *Download) CheckErrorHash(reference []byte) bool {
	return bytes.Compare(this.hash, reference) == 0
}

func (this *Download) IsFinished() bool {
	return uint64(this.received_index * DATA_BLOCK_SIZE) >= this.size
}

// moves temporary "partial" download to the final destination
// deletes info files
func (this *Download) FinishDownload() {
	err := os.Rename(this.partial_file_path, this.destination_file_path)
	CheckError(err)
	err = os.Remove(this.partial_file_path + ".info")
	CheckError(err)
}

// creates get packet for requested_index + 1 (use encode function from packets.go)
func (this *Download) CreateNextGet() []byte {

	this.requested_index += 1
	get := EncodeGet(this.requested_index)
	log.Println(this.requested_index)
	return get
}

// checks if partial download already exists
func CheckIfPartiallyDownloaded(server string, port int, rid string) (bool, string) {
	info_file_path := fmt.Sprintf("./.pft/%s@%d/%s.info", server, port, strings.Replace(rid, "/", "-", -1))
	_, err := os.Stat(info_file_path)
	exists := err == nil
	return exists, info_file_path
}

// initiates the download: creates the partial download file, initializes variables
func InitDownload(server string, port int, rid string, storage_dir string) *Download {
	server_dir := fmt.Sprintf("./.pft/%s@%d", server, port)
	err := os.MkdirAll(server_dir, 0777)
	CheckError(err)

	flat_rid := strings.Replace(rid, "/", "-", -1)
	partial_file_path := fmt.Sprintf("%s/%s", server_dir, flat_rid)

	destination_file_path, err := getDestinationFilePath(storage_dir, rid)
	CheckError(err)

	download_info := Download{
		server:                server,
		port:                  port,
		rid:                   rid,
		size:                  0,
		hash:                  nil,
		received_index:        0,
		requested_index:       0,
		partial_file_path:     partial_file_path,
		destination_file_path: destination_file_path,
	}

	download_info.SavePartialDownloadInfo()
	file, err := os.Create(download_info.partial_file_path)
	CheckError(err)
	file.Close()

	return &download_info
}

func getDestinationFilePath(storage_dir string, rid string) (string, error) {
	path := ""
	if strings.HasPrefix(rid, "file:") {
		filename := rid[5:len(rid)]
		path = fmt.Sprintf("%s/%s", storage_dir, filename)
	} else if rid == "file-list" {
		path = fmt.Sprintf("./.pft/file-list")
	} else {
		return path, errors.New("Unknown resource type")
	}

	err := os.MkdirAll(filepath.Dir(path), 0777)
	return path, err
}
