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

func runServer(port int){
    //defer wg.Done()
    local_addr, err := net.ResolveUDPAddr("udp", "127.0.0.1:" + strconv.Itoa(port))
    pft.CheckError(err)
    ChangeDir1(local_addr)

    peer := pft.MakePeer(local_addr, nil) // accept packets from any remote
    peer.Run()
}

func runClient(){
    for true {
        reader := bufio.NewReader(os.Stdin)
        fmt.Print("Request File: ")
        file_name, _ := reader.ReadString('\n')
        file_name = file_name[len(file_name) - 8:7] //TODO: Change to use all strings, not only seven character strings e.g. uno.txt
        fmt.Print("file_name: ")
        fmt.Println(file_name)

        fmt.Print("At port of localhost: ")
        request_port, _ := reader.ReadString('\n')
        fmt.Println(request_port)

        port, err := strconv.Atoi(request_port[len(request_port) - 5:4])
        local_addr, err := net.ResolveUDPAddr("udp", "127.0.0.1:0")
        pft.CheckError(err)

        server := fmt.Sprintf("%s:%d", "localhost", port)
        server_addr, err := net.ResolveUDPAddr("udp", server)
        pft.CheckError(err)

        peer := pft.MakePeer(local_addr, server_addr)

        peer.Download("file:" + file_name, server_addr)
        peer.Run()
    }
}

func main() {
    port := flag.Int("p", 4455, "ownPublishPortArg")
    flag.Parse()

    fmt.Println("Number of CPUs: ", runtime.NumCPU())
    runtime.GOMAXPROCS(runtime.NumCPU())

    var wg sync.WaitGroup
    wg.Add(2)

    fmt.Println("Starting Go Routines")

    go runServer(*port)
    go runClient()

    fmt.Println("Waiting to finish...")
    wg.Wait()

    fmt.Println("\n...terminating")
}