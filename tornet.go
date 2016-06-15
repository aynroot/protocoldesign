package main

import (
    "fmt"
    "os"
    "path/filepath"
    "net"
    "protocoldesign/pft"
    "log"
    //"path/filepath"
    "strconv"
)

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

type Chunk struct {
    filename   string
    chunkindex uint64
    hash       []byte
}

func split(filename string, nodes int64) {

    var MEGABYTE int64 = (1024 * 1024)

    // Open file for reading
    file, err := os.Open(filename)
    if err != nil {
        log.Fatal(err)
    }
    defer file.Close()

    fileInfo, err := os.Stat(filename)
    if err != nil {
        log.Fatal(err)
    }

    // name of the file
    name := fileInfo.Name()

    var sizeB, sizeM, parts, max, chunksize int64

    sizeB = fileInfo.Size()
    sizeM = sizeB / MEGABYTE

    if ((sizeM / nodes) > 1) {
        max = (sizeM / nodes)
    } else {
        max = 1
    }

    if (max < 100) {
        chunksize = max // Size in MB
    } else {
        chunksize = 100 // Size in MB
    }

    chunkoffset := MEGABYTE * chunksize  // limit to read

    // Divide depending on the megabytes

    os.MkdirAll(name, 0777)

    parts = (sizeM / chunksize)

    for i := 0; int64(i) < parts; i++ {

        newFile, err := os.Create(name + "/" + name + ".part" + strconv.Itoa(i))
        if err != nil {
            log.Fatal(err)
        }
        newFile.Close()

        //byteSlice := make([]byte, chunkoffset)
        //
        //// Read chunk to be split
        //bytesRead, err := file.Read(byteSlice)
        //if err != nil {
        //    log.Fatal(err)
        //}

    }

    byteSlice := make([]byte, chunkoffset)
    bytesRead, err := file.Read(byteSlice)
    if err != nil {
        log.Fatal(err)
    }
    log.Printf("Number of bytes read: %d\n", bytesRead)
    log.Printf("Data read: %s\n", byteSlice)
    log.Println("Size Bytes: ", sizeB)
    log.Println("Size MB: ", sizeM)
    log.Println("Nodes: ", nodes)
    log.Println("Chunk Size: ", chunksize)
    log.Println("Chunk Offset: ", chunkoffset)

    return
}
