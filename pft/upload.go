package pft

import (
    "golang.org/x/crypto/sha3"
    "fmt"
    "io/ioutil"
    "os"
)

// file path is absolute
func GetFileDataBlock(file_path string, index uint32) []byte {
    f, err := os.Open(file_path)
    check(err)

    o2, err := f.Seek(int64(index), 0)
    check(err)
    dataBlock := make([]byte, DATABLOCKSIZE)        //Todo: Check for end of file error
    n1, err := f.Read(dataBlock)
    check(err)

    fmt.Print("\n")
    fmt.Printf("%d bytes @ %d: %s\n", n1, o2, string(dataBlock))

    return dataBlock
}

func GetFileHash(file_path string) []byte {

    //read file
    file, err := ioutil.ReadFile(file_path)
    check(err)      //Todo: What to do on error (file not available)?

    hash := sha3.Sum256(file)
    var truncatedHash = hash[:16]

    fmt.Print("\n")
    fmt.Print(string(file))
    fmt.Print("\n")
    fmt.Print("Truncated Hash ", truncatedHash)
    fmt.Print("\n")
    return truncatedHash
}

// dir_path contains the path to the directory of which files are served, those are to be listed in the file-list
func GetFileListDataBlock(dir_path string, index uint32) []byte {
    var files_string string
    files, _ := ioutil.ReadDir(dir_path)
    for _, f := range files {
        files_string += f.Name()        //Todo: Add some separator? Add all filenames are presented, index and blocksize could be used to reduce the process
    }
    var files_array = []byte(files_string)
    //Check out of bounds
    if(int(index+DATABLOCKSIZE-1) < len(files_array)){
        return nil      //TODO: Throw an error (out of bounds)
    }
    fmt.Print("\n")
    var truncatedFileListDataBlock = files_array[int(index)*DATABLOCKSIZE:int(index)*DATABLOCKSIZE+DATABLOCKSIZE]
    fmt.Print(string(truncatedFileListDataBlock))
    return truncatedFileListDataBlock
}

func GetFileListHash(dir_path string, index uint32) []byte {
    var files_string string
    files, _ := ioutil.ReadDir(dir_path)
    for _, f := range files {
        files_string += f.Name()        //Todo: Add some separator?
    }

    fmt.Println(files_string)
    var files_array = []byte(files_string)
    //Check out of bounds
    if(int(index+DATABLOCKSIZE-1) < len(files_array)){
        return nil      //TODO: Throw an error (out of bounds)
    }
    files_array = files_array[int(index)*DATABLOCKSIZE:int(index)*DATABLOCKSIZE+DATABLOCKSIZE]

    hash := sha3.Sum256(files_array)
    var truncatedHash = hash[:16]


    return truncatedHash
}

func TestUploadFunc() {
    //pft.GetFileHash("/tmp/dat/test.txt")
    pwd, _ := os.Getwd()


    //Get File Hash
    fmt.Print("Get File Hash\n")
    GetFileHash(pwd+"/tmp/dat/test.txt")

    //Get File DataBlock
    fmt.Print("Get File DataBlock\n")
    GetFileDataBlock(pwd+"/tmp/dat/test.txt", 0)
    GetFileDataBlock(pwd+"/tmp/dat/test.txt", 1)

    //Get File List Hash
    fmt.Print("Get File List Hash\n")
    GetFileListHash(pwd+"/tmp/dat/", 0)

    //Load File List DataBlock
    fmt.Print("Load File List DataBlock\n")
    GetFileListDataBlock(pwd+"/tmp/dat/", 0)
    GetFileListDataBlock(pwd+"/tmp/dat/", 10)

}