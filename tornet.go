package main

import (
    "fmt"
    "os"
    "path/filepath"
    "net"
    "protocoldesign/pft"
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
            path = filepath.ToSlash(path[len(files_dir) + 1:])
            files_list = append(files_list, path)
        }
        return nil
    })

    fmt.Println(files_dir)
    fmt.Println(files_list)
    fmt.Println(nodes_list)

    os.Exit(1)

    // bind to a random port
    local_addr, err := net.ResolveUDPAddr("udp", "127.0.0.1:0")
    pft.CheckError(err)
    peer := pft.MakePeer(local_addr, nil) // accept packets from any remote

    chunk_path := "chunk_path"

    for _, node := range nodes_list {
        peer.Upload("file:" + chunk_path, node)
        peer.Run()
    }


    // establish connection to every node

    //} else {
    //    // client mode: bind to random port
    //    local_addr, err := net.ResolveUDPAddr("udp", "127.0.0.1:0")
    //    pft.CheckError(err)
    //
    //    server := fmt.Sprintf("%s:%d", flag.Arg(0), *portArg)
    //    server_addr, err := net.ResolveUDPAddr("udp", server)
    //    pft.CheckError(err)
    //
    //    peer := pft.MakePeer(local_addr, server_addr) // accept only packets from server_addr
    //
    //    if *uploadArg != "" {
    //        // upload mode
    //        peer.Upload("file:" + *uploadArg, server_addr)
    //    } else {
    //        // download mode
    //
    //        resource := "file-list" // download file list if no file specified
    //        if *downloadArg != "" {
    //            resource = "file:" + *downloadArg
    //        }
    //        peer.Download(resource, server_addr)
    //    }
    //
    //    peer.Run()
    //}

}
