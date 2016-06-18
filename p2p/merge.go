package p2p

import (
	"log"
	"protocoldesign/pft"
	"protocoldesign/tornet"
	"os"
	"bytes"
)

//TODO receive the path to the torrent OR the torrent
func MergeFile(Chunks []string) bool {

	// Extract data
	torrent_file_name := "torrent-files/test.pdf.torrent"
	torrent_file := tornet.Torrent{}
	torrent_file.Read(torrent_file_name)
	FileHash := torrent_file.FileHash

	location := "pft-files/" + torrent_file.FilePath

	for _, Chunk := range Chunks {

		//Open (if doesn't exists, create) file in append mode.
		merged_file, err := os.OpenFile(location, os.O_CREATE|os.O_APPEND|os.O_WRONLY , 0744)
		pft.CheckError(err)
		defer merged_file.Close()

		// Checking if file exists and getting its data
		chunk_info, err := os.Stat(Chunk)
		if err != nil {
			if os.IsNotExist(err) {
				log.Fatal("File "+ Chunk +" does not exist.")
				return false
			}
		}

		// Opening chunk
		chunk_file, err := os.Open(Chunk)
		pft.CheckError(err)
		defer chunk_file.Close()

		log.Println("Extracting: " + chunk_info.Name())

		chunk_size := chunk_info.Size()
		chunk_data := make([]byte, chunk_size)

		bytes_read, err := chunk_file.Read(chunk_data)
		pft.CheckError(err)

		log.Printf("Writing the following bytes: %d\n", bytes_read)

		merged_file.Write(chunk_data)
		defer merged_file.Close()
	}
	Mergedhash := tornet.CalcHash(location)

	if bytes.Equal(Mergedhash, FileHash) {

		log.Println("File reconstructed successfuly", location)
		return true

	} else {

		log.Println("The file was corrupt, and has been deleted.")
		err := os.Remove(location)
		pft.CheckError(err)
		return false
	}

}
