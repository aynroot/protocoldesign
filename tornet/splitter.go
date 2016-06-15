package tornet

import (
    "os"
    "protocoldesign/pft"
    "log"
    "strconv"
)

func SplitInChunks(filename string, n_nodes int64) {
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