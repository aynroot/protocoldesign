package p2p

import (
    "net"
    "protocoldesign/pft"
    "fmt"
    "strconv"
    "sort"
    "math/rand"
    "os"
    "protocoldesign/tornet"
    "bytes"
    "log"
    "time"
)

func GetChunksFromTracker(port int, tracker_addr *net.UDPAddr){
    local_addr, err := net.ResolveUDPAddr("udp", "127.0.0.1:" + strconv.Itoa(port))
    pft.ChangeDir(local_addr)
    pft.CheckError(err)

    // continuously accept all chunks that tracker tries to push
    peer := pft.MakePeer(local_addr, tracker_addr)
    for true {

        // Benchmarking starts to track download time.
        start := time.Now()

        peer.Run()

        elapsed := time.Since(start)
        log.Printf("Download took : %s", elapsed)

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

func downloadChunk(chunk pft.Chunk) pft.Chunk {
    local_addr, err := net.ResolveUDPAddr("udp", "127.0.0.1:0")
    pft.CheckError(err)

    random_node := chunk.Nodes[rand.Intn(len(chunk.Nodes))]
    server_addr, err := net.ResolveUDPAddr("udp", random_node)
    pft.CheckError(err)

    peer := pft.MakePeer(local_addr, server_addr)
    peer.Download("file:" + chunk.FilePath, server_addr)
    peer.Run()

    local_chunk := pft.Chunk{
        FilePath: chunk.FilePath,
        ChunkIndex: chunk.ChunkIndex,
        Hash: chunk.Hash,
    }
    return local_chunk
}

func notifyTracker(tracker_ip string, info_byte byte, chunk_rid string) {
    local_addr, err := net.ResolveUDPAddr("udp", "127.0.0.1:0")
    pft.CheckError(err)

    server_addr, err := net.ResolveUDPAddr("udp", tracker_ip)
    pft.CheckError(err)

    peer := pft.MakePeer(local_addr, server_addr)
    peer.SendChunkNotification(chunk_rid, info_byte, server_addr)
    peer.Run()
}

func registerAtTracker(tracker_ip string, chunk_rid string) {
    notifyTracker(tracker_ip, 1, chunk_rid)
}

func unregisterAtTracker(tracker_ip string, chunk_rid string) {       //TODO: Call function when chunk is not available at a specific node anymore
    notifyTracker(tracker_ip, 0, chunk_rid)
}

func RunDownloader(torrent pft.Torrent, port int) {
    file := DownloadedFile{
        file_path: torrent.FilePath,
        file_hash: torrent.FileHash,
    }

    // iterate in order of chunk indices
    var indices []int
    for index_str := range torrent.ChunksMap {
        index_int, _ := strconv.Atoi(index_str)
        indices = append(indices, index_int)
    }
    sort.Ints(indices)
    for _, index := range indices {
        chunk := torrent.ChunksMap[strconv.Itoa(index)]

        fmt.Printf("downloading chunk #%d with path %s\n", chunk.ChunkIndex, chunk.FilePath)

        var local_chunk pft.Chunk;
        if _, err := os.Stat("pft-files/" + chunk.FilePath); !os.IsNotExist(err) {
            if bytes.Equal(tornet.CalcHash("pft-files/" + chunk.FilePath), chunk.Hash) {
                log.Println("chunk already exists")
                local_chunk = chunk
            } else {
                local_chunk = downloadChunk(chunk)
            }
        } else {
            local_chunk = downloadChunk(chunk)
        }

        chunk_rid := "127.0.0.1:" + strconv.Itoa(port) + "/" + chunk.FilePath
        registerAtTracker(torrent.TrackerIP, chunk_rid)
        file.local_chunks = append(file.local_chunks, local_chunk)

        fmt.Println("saved chunk #", chunk.ChunkIndex)
    }
    MergeFile(file)
    fmt.Printf("Saved file %s on disk\n\n", file.file_path)
}