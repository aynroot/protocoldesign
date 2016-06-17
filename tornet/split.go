package tornet

import (
    "os"
    "protocoldesign/pft"
    "log"
    "strconv"
    "math"
    "golang.org/x/crypto/sha3"
)

type Chunk struct {
    FilePath   string `json:"file_path"`
    ChunkIndex uint64 `json:"chunk_index"`
    Hash       []byte `json:"hash"`
}

func SplitInChunks(file_path string, n_nodes int64) []Chunk {
    const MEGABYTE int64 = 1024 * 1024

    file, err := os.Open(file_path)
    pft.CheckError(err)
    defer file.Close()

    file_info, err := os.Stat(file_path)
    pft.CheckError(err)

    file_name := file_info.Name()
    os.MkdirAll("pft-files/_" + file_name, 0744)
    log.Println("File: ", file_path)

    size := file_info.Size()
    chunk_size := int64(math.Ceil(float64(size) / float64(n_nodes)))
    chunk_size = pft.Min(pft.Max(1 * MEGABYTE, chunk_size), 100 * MEGABYTE)
    log.Println("File size (bytes): ", size)
    log.Printf("Chunk size: %d (%.2f Mb)\n", chunk_size, float64(chunk_size) / float64(MEGABYTE))

    n_chunks := size / chunk_size
    if size % chunk_size != 0 {
        n_chunks += 1
    }
    log.Println("Number of chunks: ", n_chunks)

    chunks := make([]Chunk, 0)
    for i := 0; int64(i) < n_chunks; i++ {
        new_chunk := createChunk("pft-files/_" + file_name + "/" + file_name + ".part" + strconv.Itoa(i), i, file, chunk_size)
        chunks = append(chunks, new_chunk)
    }
    return chunks
}

func createChunk(location string, index int, file *os.File, size int64) Chunk {
    chunk_data := make([]byte, size)
    bytes_read, err := file.Read(chunk_data)
    pft.CheckError(err)
    log.Printf("Number of bytes read: %d\n", bytes_read)
    log.Printf("Wrote to file: %s\n", location)

    new_file, err := os.Create(location)
    pft.CheckError(err)

    new_file.Write(chunk_data)
    new_file.Close()

    hash := sha3.Sum256(chunk_data)
    chunk := Chunk{
        ChunkIndex: uint64(index),
        FilePath: location[len("pft-files/"):],
        Hash: hash[:],
    }

    return chunk
}