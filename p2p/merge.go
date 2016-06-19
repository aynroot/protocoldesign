package p2p

import (
    "log"
    "protocoldesign/pft"
    "protocoldesign/tornet"
    "os"
    "bytes"
    "fmt"
)

type DownloadedFile struct {
    file_path    string
    file_hash    []byte
    local_chunks []tornet.Chunk
}

func Test() {
    // important: put directories  127.0.0.1_4466 and 127.0.0.1_4467 to pft-files fot it to work
    torrent_file_name := "torrent-files/test.pdf.torrent"
    torrent := tornet.Torrent{}
    torrent.Read(torrent_file_name)

    fmt.Println(torrent.FileHash)

    file := DownloadedFile{
        file_path: torrent.FilePath,
        file_hash: torrent.FileHash,
    }

    chunk1 := tornet.Chunk{
        ChunkIndex: 0,
        FilePath: "127.0.0.1_4466/pft-files/_test.pdf/test.pdf.part0",
    }
    chunk2 := tornet.Chunk{
        ChunkIndex: 1,
        FilePath: "127.0.0.1_4467/pft-files/_test.pdf/test.pdf.part1",
    }
    file.local_chunks = append(file.local_chunks, chunk1)
    file.local_chunks = append(file.local_chunks, chunk2)

    MergeFile(file)
}

func MergeFile(file DownloadedFile) bool {
    file_hash := file.file_hash
    location := "pft-files/" + file.file_path
    fmt.Println(location)

    err := os.Remove(location)
    defer pft.CheckError(err)

    // open (if doesn't exists, create) file in append mode.
    merged_file, err := os.OpenFile(location, os.O_CREATE | os.O_APPEND | os.O_WRONLY, 0644)
    pft.CheckError(err)

    for _, chunk := range file.local_chunks {
        // checking if file exists and getting its data
        chunk_file_path := "pft-files/" + chunk.FilePath
        chunk_info, err := os.Stat(chunk_file_path)
        if err != nil {
            if os.IsNotExist(err) {
                log.Fatal("File " + chunk_file_path + " does not exist.")
                return false
            }
        }

        // opening chunk
        chunk_file, err := os.Open(chunk_file_path)
        pft.CheckError(err)
        log.Println("Extracting: " + chunk_file_path)

        chunk_size := chunk_info.Size()
        chunk_data := make([]byte, chunk_size)

        bytes_read, err := chunk_file.Read(chunk_data)
        pft.CheckError(err)
        chunk_file.Close()
        log.Printf("have read %d bytes\n", bytes_read)

        merged_file.Write(chunk_data)
    }
    merged_file.Close()

    // original
    hash := tornet.CalcHash("../torrent-files/test.pdf")
    fmt.Println(hash)

    // new one
    merged_hash := tornet.CalcHash(location)
    fmt.Println(merged_hash)

    if bytes.Equal(merged_hash, file_hash) {
        log.Println("File reconstructed successfuly: ", location)
        return true
    } else {
        log.Println("The file was corrupt, and has been deleted.")
        err := os.Remove(location)
        pft.CheckError(err)
        return false
    }

}
