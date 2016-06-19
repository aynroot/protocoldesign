package p2p

import (
    "net"
    "protocoldesign/pft"
    "protocoldesign/tornet"
    "fmt"
    "strconv"
    "strings"
)

func GetChunksFromTracker(port int, tracker_addr *net.UDPAddr){
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

func RunLocalServer(port int) {
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

func notifyTracker(tracker_ip string, info_byte []byte, chunk tornet.Chunk) {
    local_addr, err := net.ResolveUDPAddr("udp", "127.0.0.1:0")
    pft.CheckError(err)

    server := fmt.Sprintf("%s:%d", "localhost", 4455)       //Please notice: Fixed Tracker Port 4455!
    server_addr, err := net.ResolveUDPAddr("udp", server)
    pft.CheckError(err)

    peer := pft.MakePeer(local_addr, server_addr) // accept only packets from server_addr
    peer.SendNotification(chunk.FilePath, info_byte, server_addr)
    peer.Run()
}

func registerAtTracker(tracker_ip string, chunk tornet.Chunk) {
    notifyTracker(tracker_ip, []byte{1}, chunk)
}
func unregisterAtTracker(tracker_ip string, chunk tornet.Chunk) {       //TODO: Call function when node is overloaded or file not available anymore
    notifyTracker(tracker_ip, []byte{0}, chunk)
}

func RunDownloader(torrent tornet.Torrent) {
    file := DownloadedFile{
        file_path: torrent.FilePath,
        file_hash: torrent.FileHash,
    }
    for _, chunk := range torrent.Chunks {
        fmt.Printf("downloading chunk #%d with path %s\n", chunk.ChunkIndex, chunk.FilePath)
        local_chunk := downloadChunk(chunk)
        registerAtTracker(torrent.TrackerIP, chunk)
        file.local_chunks = append(file.local_chunks, local_chunk)

        fmt.Println("saved chunk #", chunk.ChunkIndex)
    }
    MergeFile(file)
    fmt.Printf("Saved file %s on disk\n\n", file.file_path)
}