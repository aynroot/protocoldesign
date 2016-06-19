package pft

import (
	"fmt"
	"os"
	"strings"
	"errors"
	"path/filepath"
	"net"
	"math/rand"
	"encoding/json"
	"io/ioutil"
)

const (
	CLOSED    = iota
	HALF_OPEN = iota
	OPEN      = iota
)

const (
	REQ      byte = 1
	REQ_ACK  byte = 2
	REQ_NACK byte = 3
	PUSH     byte = 4
	PUSH_ACK byte = 5
	GET      byte = 6
	DATA     byte = 7
	RST      byte = 8
	CNTF	 byte = 9
)

const DATA_BLOCK_SIZE = 491
const UDP_BUFFER_SIZE = 1024

const DEADLINE_SECONDS = 2

func CheckError(err error) {
    if err != nil {
        fmt.Println("Error: ", err)
        os.Exit(1)
    }
}

func GetFileDir() string {
    wd, err := os.Getwd()
    CheckError(err)
    file_dir := filepath.Join(wd, "pft-files")
    err = os.MkdirAll(file_dir, 0744)
    CheckError(err)
    return file_dir
}

func GetFileFromRID(rid string) string {
	if strings.HasPrefix(rid, "file:") {
		return rid[5:]
	} else if rid == "file-list" {
		return rid
	} else {
		CheckError(errors.New("invalid resource type"))
	}
	return ""
}

func Min(x, y int64) int64 {
	if x < y {
		return x
	}
	return y
}

func Max(x, y int64) int64 {
	if x > y {
		return x
	}
	return y
}


func generateRandomString() string {
	var letterRunes = []rune("abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ")
	b := make([]rune, 10)
	for i := range b {
		b[i] = letterRunes[rand.Intn(len(letterRunes))]
	}
	return string(b)
}

func ChangeDir(local_addr *net.UDPAddr) {
	dir := strings.Replace(local_addr.String(), ":", "_", -1)
	if (local_addr.Port == 0) {
		rand_string := generateRandomString()
		dir = dir + "_" + rand_string
	}

	err := os.MkdirAll(dir + "/pft-files", 0755)
	CheckError(err)
	err = os.Chdir(dir)
	CheckError(err)

	fmt.Println("current dir is: " + dir)
}

type Chunk struct {
	FilePath   string `json:"file_path"`
	ChunkIndex uint64 `json:"chunk_index"`
	Hash       []byte `json:"hash"`
	Nodes      []string `json:"nodes"`
}

type Torrent struct {
	TrackerIP string `json:"tracker_ip"`
	FilePath  string `json:"file_path"`
	FileHash  []byte `json:"file_hash"`
	ChunksMap map[string]Chunk `json:"chunks_map"` // ket is chunk index (int as string)
}

func (this *Torrent) Write(parent_dir string) string {
	data, err := json.Marshal(this)
	CheckError(err)

	file_path := filepath.Join(parent_dir, this.FilePath + ".torrent")
	os.MkdirAll(filepath.Dir(file_path), 0755)
	err = ioutil.WriteFile(file_path, data, 0755)
	CheckError(err)
	return file_path
}

func (this *Torrent) Read(file_path string) {
	data, err := ioutil.ReadFile(file_path)
	CheckError(err)

	err = json.Unmarshal(data, this)
	CheckError(err)
}