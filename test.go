package main

import (
    "net"
    "protocoldesign/pft"
    "os"
    "strings"
    "fmt"
    "math/rand"
)

func generateRandomString1() string {
    var letterRunes = []rune("abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ")
    b := make([]rune, 10)
    for i := range b {
        b[i] = letterRunes[rand.Intn(len(letterRunes))]
    }
    return string(b)
}

// TODO: get rid of this shit

func ChangeDir1(local_addr *net.UDPAddr) {
    dir := strings.Replace(local_addr.String(), ":", "_", -1)
    if (local_addr.Port == 0) {
        rand_string := generateRandomString1()
        dir = dir + "_" + rand_string
    }

    err := os.MkdirAll(dir + "/pft-files", 0755)
    pft.CheckError(err)
    err = os.Chdir(dir)
    pft.CheckError(err)

    fmt.Println("current dir is: " + dir)
}

func main() {
    local_addr, err := net.ResolveUDPAddr("udp", "127.0.0.1:" + os.Args[1])
    pft.CheckError(err)
    ChangeDir1(local_addr)

    peer := pft.MakePeer(local_addr, nil) // accept packets from any remote
    peer.Run()
}