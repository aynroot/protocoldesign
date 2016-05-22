package pft

import (
	"os"
	"golang.org/x/crypto/sha3"
	"strings"
    "errors"
	"path/filepath"
    "io"
)

type UploadFile struct {
    handle *os.File
    hash []byte
}

var file_cache = make(map[string]UploadFile)

func GetUploadFile(path string) (UploadFile, error) {
    upload_file, ok := file_cache[path]
    if ok {
        // fast path: return from cache
        return upload_file, nil
    }

    // open & lock file
    file, err := os.Open(path)
    if err != nil {
        return UploadFile{}, err
    }

    // calculate hash
    hasher := sha3.New256()
    io.Copy(hasher, file)

    // cache file
    upload_file = UploadFile{file, hasher.Sum(nil)}
    file_cache[path] = upload_file
    return upload_file, nil
}

func CloseUploadFile(path string) {
    upload_file, ok := file_cache[path]
    if !ok {
        return
    }

    upload_file.handle.Close()
    delete(file_cache, path)
}


func GetDataBlock(rid string, index uint32) ([]byte, error) {
    if strings.HasPrefix(rid, "file:") {
        return getFileDataBlock(filepath.Join(GetFileDir(), rid[5:]), index), nil
    } else {
        return getFileListDataBlock(GetFileDir(), index)
    }
}

// file path is absolute
func getFileDataBlock(file_path string, index uint32) []byte {
    upload_file, err := GetUploadFile(file_path)
	CheckError(err)

	_, err = upload_file.handle.Seek(int64(index  * DATA_BLOCK_SIZE), 0)
	CheckError(err)
	data_block := make([]byte, DATA_BLOCK_SIZE)
	n, err := upload_file.handle.Read(data_block)
	CheckError(err)

	return data_block[:n]
}


func GetFileHash(file_path string) []byte {
	upload_file, err := GetUploadFile(file_path)
	CheckError(err)

	return upload_file.hash
}

func getFileList(storage_dir string) []byte {

	var file_names []string
    filepath.Walk(storage_dir, func(path string, f os.FileInfo, err error) error {
        if !f.IsDir() && !strings.HasSuffix(path, ".part") && !strings.HasSuffix(path, "file-list"){
            path = filepath.ToSlash(path[len(storage_dir) + 1:])
            file_names = append(file_names, path)
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