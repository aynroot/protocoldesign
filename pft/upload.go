package pft

import (
	"os"
	"io/ioutil"

	"golang.org/x/crypto/sha3"
	"strings"
	"math"
    "errors"
    "fmt"
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
        if !strings.HasSuffix(f.Name(), ".part") {
            file_names = append(file_names, f.Name())
        }
	}
	files_string := strings.Join(file_names, "\n")
	files_array := []byte(files_string)

	//TODO: what if file-list changed?
	return files_array
}

// dir_path contains the path to the directory of which files are served, those are to be listed in the file-list
func getFileListDataBlock(storage_dir string, index uint32) ([]byte, error) {
    file_list := getFileList(storage_dir)

    if (int(index - 1) * DATA_BLOCK_SIZE >= len(file_list)) {
        return nil, errors.New("index out of bounds")
    }

    data_block := file_list[
        int(index - 1) * DATA_BLOCK_SIZE :
        int(math.Min(float64(index) * DATA_BLOCK_SIZE, float64(len(file_list))))]
	return data_block, nil
}

func GetFileListSizeAndHash(storage_dir string) (uint64, []byte) {
	file_list := getFileList(storage_dir)
	hash := sha3.Sum256(file_list)
	return uint64(len(file_list)), hash[:]
}