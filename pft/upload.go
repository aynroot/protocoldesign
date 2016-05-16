package pft

import (
	"os"
	"io/ioutil"

	"golang.org/x/crypto/sha3"
	"strings"
	"math"
)

// TODO: This is horribly slow, don't open and close the file on every call. Cache the files and close them after a timeout

// file path is absolute
func GetFileDataBlock(file_path string, index uint32) []byte {
	f, err := os.Open(file_path)
	CheckError(err)
	defer f.Close()

	_, err = f.Seek(int64((index - 1) * DATA_BLOCK_SIZE), 0)
	CheckError(err)
	data_block := make([]byte, DATA_BLOCK_SIZE)
	n, err := f.Read(data_block)
	CheckError(err)

	return data_block[:n]
}

func GetFileHash(file_path string) []byte {
	file_data, err := ioutil.ReadFile(file_path)
	CheckError(err)

	hash := sha3.Sum256(file_data)
	return hash[:]
}

func getFileList(storage_dir string) []byte {
	files, _ := ioutil.ReadDir(storage_dir)

	var file_names []string
	for _, f := range files {
		file_names = append(file_names, f.Name())
	}
	files_string := strings.Join(file_names, "\n")
	files_array := []byte(files_string)

	//TODO: what if file-list changed?
	return files_array
}

// dir_path contains the path to the directory of which files are served, those are to be listed in the file-list
func GetFileListDataBlock(storage_dir string, index uint32) []byte {
	file_list := getFileList(storage_dir)
	data_block := file_list[
        int(index - 1) * DATA_BLOCK_SIZE :
        int(math.Min(float64(index) * DATA_BLOCK_SIZE, float64(len(file_list))))] // float ?!
	return data_block
}

func GetFileListSizeAndHash(storage_dir string) (uint64, []byte) {
	file_list := getFileList(storage_dir)
	hash := sha3.Sum256(file_list)
	return uint64(len(file_list)), hash[:]
}