package tornet

import (
    "os"
    "protocoldesign/pft"
    "log"
    "strconv"
    "math"
)

func SplitInChunks(file_path string, n_nodes int64) {
    const MEGABYTE int64 = 1024 * 1024

    file, err := os.Open(file_path)
    pft.CheckError(err)
    defer file.Close()

    file_info, err := os.Stat(file_path)
    pft.CheckError(err)

    file_name := file_info.Name()
    os.MkdirAll(file_name, 0744)
    log.Println("File: ", file_path)

    size := file_info.Size()
    chunk_size := int64(math.Ceil(float64(size) / float64(n_nodes)))
    chunk_size = pft.Min(pft.Max(1 * MEGABYTE, chunk_size), 100 * MEGABYTE)
    log.Println("File size (bytes): ", size)
    log.Printf("Chunk size: %d (%.2f Mb)\n", chunk_size, float64(chunk_size) / float64(MEGABYTE))

    n_chunks := size / chunk_size
    tail_size := size % chunk_size
    log.Println("Number of full chunks: ", n_chunks)
    if tail_size != 0 {
        log.Println("Tail chunk size: ", tail_size)
    }

    for i := 0; int64(i) < n_chunks; i++ {
        writeChunk(file_name + "/" + file_name + ".part" + strconv.Itoa(i), file, chunk_size)
    }

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