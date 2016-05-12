package pft

import (
    "fmt"
    "os"
    "log"
    "strings"

    "io/ioutil"
    "encoding/gob"
    "bytes"
)


// this packet deals with resource downloads
// a file format for storing partial downloads needs to be developed
// a partial download file needs to save the rid, the size, the hash and the already downloaded data of the resource

type Download struct {
    server string
    port int
    rid string
    size uint64
    hash []byte
    received_index uint32
    requested_index uint32
    local_file_path string
}

func check(e error) {
    if e != nil {
        log.Fatal("Err: " + e.Error())
        os.Exit(0)
    }
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
    encoder.Encode(this.local_file_path)

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
    decoder.Decode(&this.local_file_path)

    return nil
}


// if index == received_index + 1: writes the payload to disk and updates received_index
// else: set requested_index = received_index
// returns true if the payload was written to disk
func (this Download) HandleDataPacket(index uint32, payload []byte) bool {
    return false
}

// saves the internal variables to a prtial download file
func (this Download) SavePartialDownload() {
    file, err := os.Create(this.local_file_path + ".info")
    check(err)
    defer file.Close()

    buffer := bytes.Buffer{}
    encoder := gob.NewEncoder(&buffer)
    err = encoder.Encode(&this)
    check(err)

    n, err := file.Write(buffer.Bytes())
    check(err)
    log.Printf("wrote %d bytes\n", n)
}

// loads the internal variables from a partial download file
func LoadPartialDownload(info_file_path string) *Download {
    data, err := ioutil.ReadFile(info_file_path)
    check(err)

    download_info := new(Download)
    buffer := bytes.NewBuffer(data)
    decoder := gob.NewDecoder(buffer)
    err = decoder.Decode(&download_info)
    check(err)

    log.Println(download_info)
    return download_info
}

func (this Download) IsFinished() bool {
    return (uint64) (this.received_index + 1) * 491 >= this.size
}

// creates get packet for requested_index + 1 (use encode function from packets.go)
func (this Download) CreateNextGet() []byte {
    return nil
}

// checks if partial download already exists
func CheckIfPartiallyDownloaded(server string, port int, rid string) (bool, string) {
    info_file_path := fmt.Sprintf("./.pft/%s:%d/%s.info", server, port, strings.Replace(rid, "/", ":", -1))
    _, err := os.Stat(info_file_path);
    exists := err == nil
    return exists, info_file_path
}



// initiates the download: creates the partial download file, initializes variables
func InitDownload(server string, port int, rid string) *Download {
    server_dir := fmt.Sprintf("./.pft/%s:%d", server, port)
    err := os.MkdirAll(server_dir, 0777)
    check(err)

    flat_rid := strings.Replace(rid, "/", ":", -1)
    local_file_path := fmt.Sprintf("%s/%s", server_dir, flat_rid)
    download_info := Download{
        server: server,
        port: port,
        rid: rid,
        size: 0,
        hash: nil,
        received_index: 0,
        requested_index: 0,
        local_file_path: local_file_path,
    }
    download_info.SavePartialDownload()

    return &download_info
}