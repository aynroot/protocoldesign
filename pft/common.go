package pft

import (
	"fmt"
	"os"
	"strings"
	"errors"
	"path/filepath"
	"net"
	"math/rand"
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

