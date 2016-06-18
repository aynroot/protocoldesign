package main

import (
    "net"
    "protocoldesign/pft"
    "protocoldesign/tornet"
    "fmt"
    "runtime"
    "sync"
    "flag"
    "strconv"
    "time"
    "math/rand"
    "strings"
    "os"
    "log"
)

type DownloadedFile struct {
    file_path    string
    local_chunks []tornet.Chunk
}

func runLocalServer(port int) {
    local_addr, err := net.ResolveUDPAddr("udp", "127.0.0.1:" + strconv.Itoa(port))
    pft.CheckError(err)
    pft.ChangeDir(local_addr)

    peer := pft.MakePeer(local_addr, nil)
    for true {
        peer.Run()
    }
}

func downloadChunk(chunk tornet.Chunk) tornet.Chunk {
    path_parts := strings.SplitN(chunk.FilePath, "/", 2)
    server_node := path_parts[0]
    file_name := path_parts[1]

    local_addr, err := net.ResolveUDPAddr("udp", "127.0.0.1:0")
    pft.CheckError(err)

    server_addr, err := net.ResolveUDPAddr("udp", server_node)
    pft.CheckError(err)

    peer := pft.MakePeer(local_addr, server_addr)
    peer.Download("file:" + file_name, server_addr)
    peer.Run()

    local_chunk := tornet.Chunk{
        FilePath: file_name,
        ChunkIndex: chunk.ChunkIndex,
        Hash: chunk.Hash,
    }
    return local_chunk
}



// TODO
func notifyTracker(tracker_ip string, chunk tornet.Chunk) {
    //var tracker_port = 4455
    local_addr, err := net.ResolveUDPAddr("udp", "127.0.0.1:0")
    pft.CheckError(err)
    //pft.ChangeDir(local_addr)

    server := fmt.Sprintf("%s:%d", "localhost", 4455)
    server_addr, err := net.ResolveUDPAddr("udp", server)
    pft.CheckError(err)

    fmt.Println(local_addr)
    fmt.Println(server_addr)

    peer := pft.MakePeer(local_addr, server_addr) // accept only packets from server_addr
    fmt.Println(peer)
    peer.SendNotification(uint32(chunk.ChunkIndex), chunk.FilePath, server_addr)
    peer.Run()
}

func mergeFile(DownloadedFile) {

}

func runDownloader(torrent tornet.Torrent) {
    file := DownloadedFile{
        file_path: torrent.FilePath,
    }
    for _, chunk := range torrent.Chunks {
        fmt.Printf("downloading chunk #%d with path %s\n", chunk.ChunkIndex, chunk.FilePath)
        local_chunk := downloadChunk(chunk)
        notifyTracker(torrent.TrackerIP, chunk)
        file.local_chunks = append(file.local_chunks, local_chunk)

        fmt.Println("saved chunk #", chunk.ChunkIndex)
    }
    mergeFile(file)
    fmt.Printf("Saved file %s on disk\n\n", file.file_path)
}

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
    go runLocalServer(*port)

    if len(flag.Args()) > 0 {
        torrent_file_name := flag.Args()[0]
        torrent_file := tornet.Torrent{}
        torrent_file.Read(torrent_file_name)
        log.Println(torrent_file)

        fmt.Println("Starting download routine")
        go runDownloader(torrent_file)
    }

    wg.Wait()
    fmt.Println("\n...terminating")
}