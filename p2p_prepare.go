package main

import (
    "net"
    "protocoldesign/pft"
    "os"
    "fmt"
    "flag"
    "time"
    "math/rand"
    "protocoldesign/p2p"
)

func main() {
    rand.Seed(time.Now().UnixNano())
    port := flag.Int("p", -1, "port to bind this node to")
    tracker := flag.String("t", "", "tracker address")
    flag.Parse()

    if *port == -1 {
        fmt.Println("Please specify port number to bind this node to.")
        os.Exit(1)
    }
    if *tracker == "" {
        fmt.Println("Please specify tracker address")
    }

    tracker_addr, err := net.ResolveUDPAddr("udp", *tracker)
    pft.CheckError(err)
    p2p.GetChunksFromTracker(*port, tracker_addr)
}