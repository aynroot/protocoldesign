package main

import (
    "fmt"
    "flag"
    "protocoldesign/pft"
    "net"
    "strconv"
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

    if !*serverModeArg && len(flag.Args()) != 1 {
        fmt.Println("need to supply exactly one target server in client mode")
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
        // server mode, bind to given port
        local_addr, err := net.ResolveUDPAddr("udp", ":" + strconv.Itoa(*portArg))
        pft.CheckError(err)

		peer := pft.MakePeer(local_addr, nil) // accept packets from any remote
        peer.Run()

    } else {
        // client mode: bind to random port
        local_addr, err := net.ResolveUDPAddr("udp", "127.0.0.1:0")
        pft.CheckError(err)

        server := fmt.Sprintf("%s:%d", flag.Arg(0), *portArg)
        server_addr, err := net.ResolveUDPAddr("udp", server)
        pft.CheckError(err)

        peer := pft.MakePeer(local_addr, server_addr) // accept only packets from server_addr
        peer.Download(resource, server_addr)
        peer.Run()
    }

}
