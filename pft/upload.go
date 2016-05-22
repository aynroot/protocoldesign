package pft

import (
	"os"
	"io/ioutil"
	"golang.org/x/crypto/sha3"
	"strings"
    "errors"
    "fmt"
	"path/filepath"
)

// TODO: This is horribly slow, don't open and close the file on every call. Cache the files and close them after a timeout

func GetDataBlock(rid string, index uint32) ([]byte, error) {
    if strings.HasPrefix(rid, "file:") {
        return getFileDataBlock(fmt.Sprintf("%s/%s", GetFileDir(), rid[5:len(rid)]), index), nil
    } else {
        return getFileListDataBlock(GetFileDir(), index)
    }
}

// file path is absolute
func getFileDataBlock(file_path string, index uint32) []byte {
	f, err := os.Open(file_path)
	CheckError(err)
	defer f.Close()

    // TODO: move logic about size and hash from peer here (merge with next func)
    // TODO: check for 'index out of bounds' when doing f.Seek
    // TODO: keep file open

	_, err = f.Seek(int64(index  * DATA_BLOCK_SIZE), 0)
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

	var file_names []string
    filepath.Walk(storage_dir, func(path string, f os.FileInfo, err error) error {
        if !f.IsDir() {
            file_names = append(file_names, path[len(storage_dir) + 1:])
        }
        return nil
    })
	files_string := strings.Join(file_names, "\n")
	files_array := []byte(files_string)

	return files_array
}

// dir_path contains the path to the directory of which files are served, those are to be listed in the file-list
func getFileListDataBlock(storage_dir string, index uint32) ([]byte, error) {
    file_list := getFileList(storage_dir)

    if index * DATA_BLOCK_SIZE >= uint32(len(file_list)) {
        return nil, errors.New("index out of bounds")
    }

    data_start := index * DATA_BLOCK_SIZE
	data_end := data_start + DATA_BLOCK_SIZE
    if data_end > uint32(len(file_list)) {
        data_end = uint32(len(file_list))
    }

    data_block := file_list[data_start : data_end]
	return data_block, nil
}

func GetFileListSizeAndHash(storage_dir string) (uint64, []byte) {
	file_list := getFileList(storage_dir)
	hash := sha3.Sum256(file_list)
	return uint64(len(file_list)), hash[:]
}