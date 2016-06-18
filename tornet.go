package main

import (

    "flag"
    "path/filepath"
    "os"
    "fmt"
    "net"
    "strconv"
    "protocoldesign/pft"
    "protocoldesign/tornet"
)

func main() {
    port := flag.Int("p", 4455, "port to listen to")
    flag.Parse()

    var files_dir string;
    var nodes_list []string;

    files_dir = flag.Args()[0]
    for _, arg := range flag.Args()[1:] {
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

    peer := pft.MakePeer(local_addr, nil)
    for _, file_path := range files_list {
        fmt.Println("Distributing file: ", file_path)
        torrent_file := tornet.DistributeFile(peer, local_addr, string(file_path), nodes_list)
        torrent_file_path := torrent_file.Write("torrent-files")
        fmt.Println("Created *.torrent file: ", torrent_file_path)
    }

    // TODO WTF IS THIS SHIT
    // server mode, bind to given port
    //local_addr1, err := net.ResolveUDPAddr("udp", "127.0.0.1:" + strconv.Itoa(4455))
    //pft.CheckError(err)
    //pft.ChangeDir(local_addr1)

    //peer1 := pft.MakePeer(local_addr1, nil) // accept packets from any remote
    //peer1.Run()
    for true {
        peer.Run()
    }
}
