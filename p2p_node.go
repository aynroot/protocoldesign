package

import (
    "time"
    "flag"
    "fmt"
    "os"
    "runtime"
    "sync"
    "protocoldesign/tornet"
    "log"
    "protocoldesign/p2p"
    "math/rand"
)

func main() {
    rand.Seed(time.Now().UnixNano())
    port := flag.Int("p", -1, "port to bind this node to")
    flag.Parse()

    if *port == -1 {
        fmt.Println("Please specify port number to bind this node to.")
        os.Exit(1)
    }
    if len(flag.Args()) > 1 {
        fmt.Println("Too many command line parameters")
        os.Exit(1)
    }

    runtime.GOMAXPROCS(runtime.NumCPU())
    var wg sync.WaitGroup
    wg.Add(2)

    fmt.Println("Starting upload routine")
    go p2p.RunLocalServer(*port)

    if len(flag.Args()) > 0 {
        torrent_file_name := flag.Args()[0]
        torrent_file := tornet.Torrent{}
        torrent_file.Read(torrent_file_name)
        log.Println(torrent_file)

        fmt.Println("Starting download routine")
        go p2p.RunDownloader(torrent_file)
    }

    wg.Wait()
    fmt.Println("\n...terminating")

}
