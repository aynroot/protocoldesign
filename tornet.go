package main

import (
    "fmt"
    "os"
    "path/filepath"
    "net"
    "protocoldesign/pft"
    "log"
    "strconv"
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

    split("pft-files/test.pdf", 2);

}

func split(filename string, n_nodes int64) {
    var MEGABYTE int64 = 1024 * 1024

    file, err := os.Open(filename)
    pft.CheckError(err)
    defer file.Close()

    file_info, err := os.Stat(filename)
    pft.CheckError(err)

    size := file_info.Size()
    chunk_size := pft.Min(pft.Max(1 * MEGABYTE, size / n_nodes), 100 * MEGABYTE)
    log.Println("Size (bytes): ", size)
    log.Println("Chunk size: ", chunk_size)

    n_chunks := size / chunk_size
    log.Println("Number of full chunks: ", n_chunks)

    file_name := file_info.Name()
    os.MkdirAll(file_name, 0744)

    for i := 0; int64(i) < n_chunks; i++ {
        writeChunk(file_name + "/" + file_name + ".part" + strconv.Itoa(i), file, chunk_size)
    }

    tail_size := size % chunk_size
    if (tail_size != 0) {
        writeChunk(file_name + "/" + file_name + ".part" + strconv.Itoa(int(n_chunks)), file, tail_size)
    }
}

func writeChunk(location string, file *os.File, size int64) {
    chunk_data := make([]byte, size)
    bytes_read, err := file.Read(chunk_data)
    pft.CheckError(err)
    log.Printf("Number of bytes read: %d\n", bytes_read)
    log.Printf("Wrote to file: %s\n", location)

    new_file, err := os.Create(location)
    pft.CheckError(err)

    new_file.Write(chunk_data)
    new_file.Close()
}
