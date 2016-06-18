package p2p

import (
    "log"
    "protocoldesign/pft"
    "protocoldesign/tornet"
    "os"
    "bytes"
    "fmt"
    "io"
)

func deepCompare(file1, file2 string) bool {
    const chunkSize = 64000

    f1s, err := os.Stat(file1)
    if err != nil {
        log.Fatal(err)
    }
    f2s, err := os.Stat(file2)
    if err != nil {
        log.Fatal(err)
    }

    if f1s.Size() != f2s.Size() {
        fmt.Println(f1s.Size())
        fmt.Println(f2s.Size())
        return false
    }

    f1, err := os.Open(file1)
    if err != nil {
        log.Fatal(err)
    }

    f2, err := os.Open(file2)
    if err != nil {
        log.Fatal(err)
    }

    for {
        b1 := make([]byte, chunkSize)
        _, err1 := f1.Read(b1)

        b2 := make([]byte, chunkSize)
        _, err2 := f2.Read(b2)

        if err1 != nil || err2 != nil {
            if err1 == io.EOF && err2 == io.EOF {
                return true
            } else if err1 == io.EOF || err2 == io.EOF {
                return false
            } else {
                log.Fatal(err1, err2)
            }
        }

        if !bytes.Equal(b1, b2) {
            return false
        }
    }
}

func Test() {
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

    // open (if doesn't exists, create) file in append mode.
    merged_file, err := os.OpenFile(location, os.O_CREATE | os.O_APPEND | os.O_WRONLY, 0744)
    pft.CheckError(err)

    for _, chunk := range file.local_chunks {
        // checking if file exists and getting its data
        chunk_info, err := os.Stat(chunk.FilePath)
        if err != nil {
            if os.IsNotExist(err) {
                log.Fatal("File " + chunk.FilePath + " does not exist.")
                return false
            }
        }

        // opening chunk
        chunk_file, err := os.Open(chunk.FilePath)
        pft.CheckError(err)
        log.Println("Extracting: " + chunk.FilePath)

        chunk_size := chunk_info.Size()
        chunk_data := make([]byte, chunk_size)

        bytes_read, err := chunk_file.Read(chunk_data)
        pft.CheckError(err)
        chunk_file.Close()
        log.Printf("have read %d bytes\n", bytes_read)

        merged_file.Write(chunk_data)
    }
    merged_file.Close()

    hash := tornet.CalcHash("tornet-files/test.pdf")
    fmt.Println(hash)
    merged_hash := tornet.CalcHash(location)
    fmt.Println(merged_hash)

    fmt.Println(deepCompare(location, "tornet-files/test.pdf"))

    if bytes.Equal(merged_hash, file_hash) {
        log.Println("File reconstructed successfuly: ", location)
        return true
    } else {
        log.Println("The file was corrupt, and has been deleted.")
        //err := os.Remove(location)
        //pft.CheckError(err)
        return false
    }

}
