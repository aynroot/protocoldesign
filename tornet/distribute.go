package tornet

import (
    "net"
    "protocoldesign/pft"
    "log"
    "fmt"
)

func DistributeFile(file_path string, nodes_list []string) {
    local_addr, err := net.ResolveUDPAddr("udp", "127.0.0.1:4455")
    pft.CheckError(err)

    peer := pft.MakePeer(local_addr, nil) // accept packets from any remote
    chunks := SplitInChunks(file_path, int64(len(nodes_list)))
    fmt.Println(chunks)
    n_nodes := len(nodes_list)

    for chunk_index := 0; chunk_index < len(chunks); chunk_index++ {
        node_index := chunk_index % n_nodes
        node := nodes_list[node_index]
        log.Printf("sending chunk #%d to the node #%d (%s)", chunk_index, node_index, node)

        node_addr, err := net.ResolveUDPAddr("udp", node)
        pft.CheckError(err)

        peer.Upload("file:" + chunks[chunk_index].file_path[len("pft-files/"):], node_addr)
        peer.Run()
    }
}
