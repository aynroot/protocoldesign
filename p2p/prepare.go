package p2p

import (
    "net"
    "protocoldesign/pft"
    "os"
    "fmt"
    "flag"
    "strconv"
    "time"
    "math/rand"
)

func getChunksFromTracker(port int, tracker_addr *net.UDPAddr){
    local_addr, err := net.ResolveUDPAddr("udp", "127.0.0.1:" + strconv.Itoa(port))
    pft.ChangeDir(local_addr)
    pft.CheckError(err)

    // continuously accept all chunks that tracker tries to push
    peer := pft.MakePeer(local_addr, tracker_addr)
    for true {
        peer.Run()
        fmt.Println("Chunk is received")
    }
}

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
    getChunksFromTracker(*port, tracker_addr)
}