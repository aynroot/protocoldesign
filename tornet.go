package main

import (
    "fmt"
    "os"
    "flag"
    "path/filepath"
    "protocoldesign/tornet"
    "net"
    "strconv"
    "protocoldesign/pft"
)

func main() {
    port := flag.Int("-p", 4455, "port to listen to")
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

    if (len(files_list) == 0) {
        fmt.Println("Please, specify files directory with the files you want to distribute.")
    }
    if (len(nodes_list) == 0) {
        fmt.Println("Please, specify p2p nodes addresses in format IP:PORT.")
    }

    fmt.Println("Files dir: ", files_dir)
    fmt.Println("Files list: ", files_list)
    fmt.Println("Nodes list: ", nodes_list)

    local_addr, err := net.ResolveUDPAddr("udp", "127.0.0.1:" + strconv.Itoa(*port))
    pft.CheckError(err)

    for _, file_path := range files_list {
        fmt.Println("Distributing file: ", file_path)
        torrent_file := tornet.DistributeFile(local_addr, string(file_path), nodes_list)
        torrent_file_path := torrent_file.Write("torrent-files")
        fmt.Println("Created *.torrent file: ", torrent_file_path)
    }
}
