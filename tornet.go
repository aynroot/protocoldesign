package main

import (
    "fmt"
    "os"
    "path/filepath"
    "net"
    "protocoldesign/pft"
)

type Chunk struct {
    filename   string
    chunkindex uint64
    hash       []byte
}

func main() {
    fmt.Println(len(os.Args), os.Args)

    var files_dir string;
    var nodes_list []string;

    files_dir = os.Args[1]
    for _, arg := range os.Args[2:] {
        nodes_list = append(nodes_list, arg);
    }

    var files_list []string;
    filepath.Walk(files_dir, func(path string, f os.FileInfo, err error) error {
        if !f.IsDir() {
            path = filepath.ToSlash(path[len(files_dir) + 1:])
            files_list = append(files_list, path)
        }
        return nil
    })

    fmt.Println(files_dir)
    fmt.Println(files_list)
    fmt.Println(nodes_list)

    os.Exit(1)

    // bind to a random port
    local_addr, err := net.ResolveUDPAddr("udp", "127.0.0.1:0")
    pft.CheckError(err)
    peer := pft.MakePeer(local_addr, nil) // accept packets from any remote


    // in progress
    chunk_path := "chunk_path"

    for _, node := range nodes_list {
        peer.Upload("file:" + chunk_path, node)
        peer.Run()
        // this won't work because Run is inifinite + we have os.Exit on the "client" side
    }
}
