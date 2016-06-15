package main

import (
    "fmt"
    "flag"
    "protocoldesign/pft"
    "net"
    "strconv"
    "os"
    "strings"
    "math/rand"
    "time"
)

func generateRandomString() string {
    var letterRunes = []rune("abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ")
    b := make([]rune, 10)
    for i := range b {
        b[i] = letterRunes[rand.Intn(len(letterRunes))]
    }
    return string(b)
}

func changeDir(local_addr *net.UDPAddr) {
    dir := strings.Replace(local_addr.String(), ":", "_", -1)
    if (local_addr.Port == 0) {
        rand_string := generateRandomString()
        dir = dir + "_" + rand_string
    }

    err := os.MkdirAll(dir + "/pft-files", 0755)
    pft.CheckError(err)
    err = os.Chdir(dir)
    pft.CheckError(err)

    fmt.Println("current dir is: " + dir)
}

func main() {
    portArg := flag.Int("t", 6222, "port to contact/listen on")
    serverModeArg := flag.Bool("s", false, "start in server mode")
    downloadArg := flag.String("d", "", "file to be downloaded")
    uploadArg := flag.String("u", "", "file to be uploaded")
    flag.Parse()

    if *serverModeArg && *downloadArg != "" {
        fmt.Println("can only download file in client mode")
        return
    }

    if !*serverModeArg && len(flag.Args()) != 1 {
        fmt.Println("need to supply exactly one target server in client mode")
        return
    }

    fmt.Println("port:", *portArg)
    fmt.Println("server mode:", *serverModeArg)
    fmt.Println("server:", flag.Args())
    rand.Seed(time.Now().UnixNano())

    if *serverModeArg {
        // server mode, bind to given port
        local_addr, err := net.ResolveUDPAddr("udp", "127.0.0.1:" + strconv.Itoa(*portArg))
        pft.CheckError(err)
        changeDir(local_addr)

        peer := pft.MakePeer(local_addr, nil) // accept packets from any remote
        peer.Run()
        fmt.Println("asaaaaaaaaaaaaaaaaaaaa")

    } else {
        // client mode: bind to random port
        local_addr, err := net.ResolveUDPAddr("udp", "127.0.0.1:0")
        pft.CheckError(err)
        changeDir(local_addr)

        server := fmt.Sprintf("%s:%d", flag.Arg(0), *portArg)
        server_addr, err := net.ResolveUDPAddr("udp", server)
        pft.CheckError(err)

        peer := pft.MakePeer(local_addr, server_addr) // accept only packets from server_addr

        if *uploadArg != "" {
            // upload mode
            peer.Upload("file:" + *uploadArg, server_addr)
        } else {
            // download mode

            resource := "file-list" // download file list if no file specified
            if *downloadArg != "" {
                resource = "file:" + *downloadArg
            }
            peer.Download(resource, server_addr)
        }

        peer.Run()
    }

}
