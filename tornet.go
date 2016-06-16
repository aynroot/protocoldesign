package main

import (
    "fmt"
    "os"
    "path/filepath"
    "protocoldesign/tornet"
    "log"
)

func main() {
    fmt.Println(len(os.Args), os.Args)

    var files_dir string;
    var nodes_list []string;

    files_dir = os.Args[1]
    for _, arg := range os.Args[2:] {
        nodes_list = append(nodes_list, arg);
    }

    var files_list []string;
    filepath.Walk(files_dir, func(path string, f os.FileInfo, err error) error {
        if !f.IsDir() {
            files_list = append(files_list, path)
        }
        return nil
    })

    fmt.Println(files_dir)
    fmt.Println(files_list)
    fmt.Println(nodes_list)

    if (len(files_list) == 0) {
        fmt.Println("Please, specify files directory with the files you want to distribute.")
    }
    if (len(nodes_list) == 0) {
        fmt.Println("Please, specify p2p nodes addresses in format IP:PORT.")
    }

    for _, file_path := range files_list {
        log.Println(file_path)
        tornet.DistributeFile(string(file_path), nodes_list)
    }
}
