package tornet

import (
    "net"
    "protocoldesign/pft"
    "fmt"
    "os"
    "golang.org/x/crypto/sha3"
    "io"
    "strings"
)

type Torrent struct {
    TrackerIP string `json:"tracker_ip"`
    FilePath  string `json:"file_path"`
    FileHash  []byte `json:"file_hash"`
    Chunks    []Chunk `json:"chunks"`
}

func CalcHash(file_path string) []byte {
    file, err := os.Open(file_path)
    pft.CheckError(err)
    defer file.Close()

    hasher := sha3.New256()
    io.Copy(hasher, file) // TODO: extra pass, consider move consequently copy when you build chunks
    return hasher.Sum(nil)
}

func DistributeFile(local_addr *net.UDPAddr, file_path string, nodes_list []string) Torrent {
    peer := pft.MakePeer(local_addr, nil)
    chunks := SplitInChunks(file_path, int64(len(nodes_list)))
    n_nodes := len(nodes_list)

    path_without_parent_dir := strings.SplitN(file_path, "/", 2)[1]
    tornet_file := Torrent{
        TrackerIP: local_addr.String(),
        FilePath: path_without_parent_dir,
        FileHash: CalcHash(file_path),
    }
    for chunk_index := 0; chunk_index < len(chunks); chunk_index++ {
        node_index := chunk_index % n_nodes
        node := nodes_list[node_index]
        fmt.Printf("sending chunk #%d to the node #%d (%s)\n", chunk_index, node_index, node)

        node_addr, err := net.ResolveUDPAddr("udp", node)
        pft.CheckError(err)

        peer.Upload("file:" + chunks[chunk_index].FilePath, node_addr)
        peer.Run()

        remote_path := node + "/" + chunks[chunk_index].FilePath
        remote_chunk := Chunk{
            FilePath: remote_path,
            ChunkIndex: chunks[chunk_index].ChunkIndex,
            Hash: chunks[chunk_index].Hash,
        }
        tornet_file.Chunks = append(tornet_file.Chunks, remote_chunk)
    }
    return tornet_file
}
