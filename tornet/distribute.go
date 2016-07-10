package tornet

import (
    "net"
    "protocoldesign/pft"
    "fmt"
    "os"
    "golang.org/x/crypto/sha3"
    "io"
    "strings"
    "strconv"
)

func CalcHash(file_path string) []byte {
    file, err := os.Open(file_path)
    pft.CheckError(err)
    defer file.Close()

    hasher := sha3.New256()
    io.Copy(hasher, file)
    return hasher.Sum(nil)
}

func DistributeFile(peer pft.Peer, local_addr *net.UDPAddr, file_path string, nodes_list []string) pft.Torrent {

    chunks := SplitInChunks(file_path, int64(len(nodes_list)))
    n_nodes := len(nodes_list)

    path_without_parent_dir := strings.SplitN(file_path, string(os.PathSeparator), 2)[1]

    tornet_file := pft.Torrent{
        TrackerIP: local_addr.String(),
        FilePath: path_without_parent_dir,
        FileHash: CalcHash(file_path),
        ChunksMap: make(map[string]pft.Chunk),
    }
    for chunk_index := 0; chunk_index < len(chunks); chunk_index++ {
        node_index := chunk_index % n_nodes
        node := nodes_list[node_index]
        fmt.Printf("sending chunk #%d to the node #%d (%s)\n", chunk_index, node_index, node)

        node_addr, err := net.ResolveUDPAddr("udp", node)
        pft.CheckError(err)

        peer.Upload("file:" + chunks[chunk_index].FilePath, node_addr)
        peer.Run()

        chunks[chunk_index].Nodes = []string{node}
        tornet_file.ChunksMap[strconv.Itoa(int(chunks[chunk_index].ChunkIndex))] = chunks[chunk_index]
    }

    return tornet_file
}
