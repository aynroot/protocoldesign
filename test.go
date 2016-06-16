package main

import (
    "net"
    "protocoldesign/pft"
    "os"
    "strings"
    "fmt"
    "math/rand"
    "runtime"
    "sync"
    "flag"
    "strconv"
    "bufio"
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
    ownPublishPortArg := flag.Int("p", 4455, "ownPublishPortArg")
    foreignLoadPortArg := flag.Int("c", 5566, "foreignLoadPortArg")
    consumeArg := flag.Bool("x", false, "start in server mode")
    downloadArg := flag.String("d", "", "file to be downloaded")
    flag.Parse()

    fmt.Println(foreignLoadPortArg)

    runtime.GOMAXPROCS(2)

    var wg sync.WaitGroup
    wg.Add(2)

    fmt.Println("Starting Go Routines")

    go func() {
        defer wg.Done()
        local_addr, err := net.ResolveUDPAddr("udp", "127.0.0.1:" + strconv.Itoa(*ownPublishPortArg))
        pft.CheckError(err)
        ChangeDir1(local_addr)


        peer := pft.MakePeer(local_addr, nil) // accept packets from any remote
        peer.Run()
    }()

    go func() {
        if *consumeArg {

            fmt.Print("Client-Part: ")
            fmt.Println(*consumeArg)

            for true {
                reader := bufio.NewReader(os.Stdin)
                fmt.Print("Request File: ")
                file_name, _ := reader.ReadString('\n')
                fmt.Println(file_name)
                fmt.Print("At port of localhost: ")
                request_port, _ := reader.ReadString('\n')
                fmt.Println(request_port)

                last3 := request_port[len(request_port) - 5:4]
                i, err := strconv.Atoi(last3)
                *foreignLoadPortArg = i

                local_addr, err := net.ResolveUDPAddr("udp", "127.0.0.1:0")
                pft.CheckError(err)

                ChangeDir1(local_addr)

                server := fmt.Sprintf("%s:%d", flag.Arg(0), *foreignLoadPortArg)
                server_addr, err := net.ResolveUDPAddr("udp", server)
                pft.CheckError(err)

                peer := pft.MakePeer(local_addr, server_addr) //accept only packets from server_addr; on nil: accept packets from any remote

                // download mode

                resource := "file-list" // download file list if no file specified
                if *downloadArg != "" {
                    resource = "file:" + *downloadArg
                }

                peer.Download(resource, server_addr)

                peer.Run()

            }
        }
    }()

    fmt.Println("Waiting To Finish")
    wg.Wait()

    fmt.Println("\nTerminating Program")

    //New thread
    //Start Client loading shit on demand
}