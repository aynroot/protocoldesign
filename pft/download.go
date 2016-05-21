package pft

import (
    "fmt"
    "log"
    "os"
    "encoding/gob"
    "errors"
    "path/filepath"
    "reflect"
)

// this packet deals with resource downloads
// a file format for storing partial downloads needs to be developed
// a partial download file needs to save the rid, the size, the hash and the already downloaded data of the resource

type Download struct {
    server          string
    port            int
    rid             string
    size            uint64
    hash            []byte
    received_index  uint32
    requested_index uint32
    local_file      *os.File
}


// returns true if the payload was written to disk
func (this Download) HandleDataPacket(index uint32, payload []byte) bool {
    if index != this.received_index + 1 {
        this.requested_index = this.received_index
        return false
    }

    n, err := this.local_file.Write(payload)
    CheckError(err)
    log.Println("wrote", n, "bytes to", this.local_file.Name())

    this.received_index = index
    return true
}

// saves the internal variables to a partial download file
func (this Download) CreatePartFile(path string) {
    file, err := os.OpenFile(path, os.O_TRUNC | os.O_WRONLY | os.O_CREATE, 644)
    CheckError(err)
    defer file.Close()

    encoder := gob.NewEncoder(file)
    encoder.Encode(this.server)
    encoder.Encode(this.port)
    encoder.Encode(this.rid)
    encoder.Encode(this.size)
    encoder.Encode(this.hash)
}

func (this Download) Close() {
    os.Remove(this.local_file.Name() + ".part")
    this.local_file.Close()
}

// loads the internal variables from a partial download file
func LoadPartFile(download_file_path string, size uint64, hash []byte, d *Download) error {
    part_file_path := download_file_path + ".part"
    part_file, err := os.OpenFile(part_file_path, os.O_RDONLY, 644)
    if err != nil {
        return err
    }
    defer part_file.Close()

    download_file, err := os.OpenFile(download_file_path, os.O_WRONLY | os.O_APPEND, 644)
    if err != nil {
        return err
    }
    download_info, err := download_file.Stat()
    if err != nil {
        download_file.Close()
        return err
    }
    if download_info.Size() % DATA_BLOCK_SIZE != 0 {
        download_file.Close()
        return errors.New("partial download file size is not divisile by data block size")
    }

    decoder := gob.NewDecoder(part_file)
    decoder.Decode(d.server)
    decoder.Decode(d.port)
    decoder.Decode(d.rid)
    decoder.Decode(d.size)
    err = decoder.Decode(d.hash)
    if err != nil {
        download_file.Close()
        return err
    }

    if d.size != size || !reflect.DeepEqual(d.hash, hash) {
        // file has changed
        download_file.Close()
        return errors.New("file has changed since part was downloaded")
    }

    d.received_index = uint32(download_info.Size() / DATA_BLOCK_SIZE)
    d.requested_index = uint32(download_info.Size() / DATA_BLOCK_SIZE)
    d.local_file = download_file

    return nil
}

func (this *Download) IsFinished() bool {
    return uint64(this.received_index * DATA_BLOCK_SIZE) >= this.size
}

// creates get packet for requested_index + 1 (use encode function from packets.go)
func (this Download) CreateNextGet() []byte {
    this.requested_index += 1
    get := EncodeGet(this.requested_index)
    log.Println("requesting", this.requested_index)
    return get
}

func (this Download) ResetGet() {
    this.requested_index = this.received_index
}

// creates a download object: either continues a partial download or creates a new one
func InitDownload(server string, port int, rid string, size uint64, hash []byte) Download {
    download_file_path := fmt.Sprintf("%s/%s", GetFileDir(), filepath.Base(rid))
    part_file_path := download_file_path + ".part"

    var d Download
    if err := LoadPartFile(download_file_path, size, hash, &d); err == nil {
        // download loaded from .part file
        return d
    }

    file, err := os.OpenFile(download_file_path, os.O_WRONLY | os.O_CREATE | os.O_TRUNC, 644)
    CheckError(err)

    d.server = server
    d.port = port
    d.rid = rid
    d.size = size
    d.hash = hash
    d.received_index = 0
    d.requested_index = 0
    d.local_file = file

    d.CreatePartFile(part_file_path)
    return d
}