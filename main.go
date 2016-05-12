package main

import (
    "fmt"
    "net"
    "os"
    "flag"
    "strconv"
    "time"

    "./pft"
    "log"
)

func main() {
    portArg := flag.Int("t", 6222, "port to contact/listen on")
    serverModeArg := flag.Bool("s", false, "start in server mode")
    fileArg := flag.String("f", "", "file to be downloaded")
    flag.Parse()

    if *serverModeArg && *fileArg != "" {
        fmt.Println("can only download file in client mode")
        return
    }

    resource := "file-list"
    if *fileArg != "" {
        resource = "file:" + *fileArg
    }

    fmt.Println("port:", *portArg)
    fmt.Println("server mode:", *serverModeArg)
    fmt.Println("server:", flag.Args())


    if *serverModeArg {
        Server(*portArg)
    } else {
        if len(flag.Args()) != 1 {
            fmt.Println("need to supply exactly one target server in client mode")
            return
        }
        Client(*portArg, flag.Args()[0], resource)
    }
}


func CheckError(err error) {
    if err != nil {
        fmt.Println("Error: ", err)
        os.Exit(0)
    }
}


func Server(port int) {
    addr, err := net.ResolveUDPAddr("udp", ":" + strconv.Itoa(port))
    CheckError(err)

    fmt.Println("listening on", addr)

    /* Now listen at selected port */
    conn, err := net.ListenUDP("udp", addr)
    CheckError(err)
    defer conn.Close()

    buf := make([]byte, 512)
    for {
        size, sender, err := conn.ReadFromUDP(buf)
        CheckError(err)
        fmt.Println("Received ", string(buf[0:size]), " from ", sender)
    }
    fmt.Println("closing server")
}


func Client(port int, server string, resource string) {
    server_addr, err := net.ResolveUDPAddr("udp", server + ":" + strconv.Itoa(port))
    CheckError(err)

    local_addr, err := net.ResolveUDPAddr("udp", "127.0.0.1:0")
    CheckError(err)

    conn, err := net.DialUDP("udp", local_addr, server_addr)
    CheckError(err)
    defer conn.Close()

    storage_dir := "./files"
    exists, info_file_path := pft.CheckIfPartiallyDownloaded(server, port, resource)
    download := new(pft.Download)
    if exists {
        download = pft.LoadPartialDownload(info_file_path)
    } else {
        download = pft.InitDownload(server, port, resource, storage_dir)
    }
    fmt.Println(download)

    testDownload(*download)

    i := 1
    for {
        conn.Write([]byte(strconv.Itoa(i)))
        i += 1
        time.Sleep(time.Second)
    }

}

func testDownload(download pft.Download) {
    // make sure to set PAYLOAD_SIZE = 5 in download.go and init download size as something small
    var i uint32;
    for i = 1; i < 10; i++ {
        download.HandleDataPacket(i, []byte{115, 111, 109, 101, 10}) // write "some\n"
        if (download.IsFinished()) {
            download.FinishDownload()
            log.Println("finished")
            break
        }
    }
}

