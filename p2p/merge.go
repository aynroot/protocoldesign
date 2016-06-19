package p2p

import (
    "bytes"
    "log"
    "os"
    "protocoldesign/pft"
    "protocoldesign/tornet"
)

type DownloadedFile struct {
    file_path    string
    file_hash    []byte
    local_chunks []pft.Chunk
}

func MergeFile(file DownloadedFile) bool {
    file_hash := file.file_hash

    location := "pft-files" + string(os.PathSeparator) + file.file_path
    log.Println("merging: ", location)

    err := os.Remove(location) // err != nil if location doesn't exist

    // open (if doesn't exists, create) file in append mode.
    merged_file, err := os.OpenFile(location, os.O_CREATE | os.O_APPEND | os.O_WRONLY, 0644)
    pft.CheckError(err)

    for _, chunk := range file.local_chunks {
        // checking if file exists and getting its data
        chunk_file_path := "pft-files" + string(os.PathSeparator) + chunk.FilePath
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

    merged_hash := tornet.CalcHash(location)
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
