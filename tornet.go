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
    "encoding/json"
    "io/ioutil"
    "log"
    "strings"
)

func saveTorrentMap(torrent_map map[string]pft.Torrent) {
    data, err := json.Marshal(torrent_map)
    pft.CheckError(err)

    err = ioutil.WriteFile("tornet.json", data, 0755)
    pft.CheckError(err)

    log.Println("torrent map is saved")
}

func initTorrentMap() map[string]pft.Torrent {
    torrent_map := make(map[string]pft.Torrent)

    data, err := ioutil.ReadFile("tornet.json")
    pft.CheckError(err)
    err = json.Unmarshal(data, &torrent_map)
    pft.CheckError(err)

    return torrent_map
}

func parseArgs(args []string) ([]string, []string) {
    var files_dir string;
    var nodes_list []string;

    files_dir = args[0]
    for _, arg := range args[1:] {
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
        os.Exit(1)
    }
    if (len(nodes_list) == 0) {
        fmt.Println("Please, specify p2p nodes addresses in format IP:PORT.")
        os.Exit(1)
    }

    return files_list, nodes_list
}

func distributeFiles(files_list []string, nodes_list []string, peer pft.Peer, local_addr *net.UDPAddr) map[string]pft.Torrent {
    torrent_map := make(map[string]pft.Torrent)
    for _, file_path := range files_list {
        fmt.Println("Distributing file: ", file_path)
        torrent := tornet.DistributeFile(peer, local_addr, string(file_path), nodes_list)
        torrent_path := torrent.Write("torrent-files")

        file_name := strings.SplitN(file_path, string(os.PathSeparator), 2)[1]
        torrent_map[file_name] = torrent
        fmt.Println("Created *.torrent file: ", torrent_path)
    }
    return torrent_map
}

func main() {
    port := flag.Int("p", 4455, "port to listen to")
    flag.Parse()

    local_addr, err := net.ResolveUDPAddr("udp", "127.0.0.1:" + strconv.Itoa(*port))
    pft.CheckError(err)

    peer := pft.MakePeer(local_addr, nil)
    var torrent_map map[string]pft.Torrent

    if len(flag.Args()) > 0 {
        // preparation process
        files_list, nodes_list := parseArgs(flag.Args())

        fmt.Println("Files list: ", files_list)
        fmt.Println("Nodes list: ", nodes_list)

        torrent_map = distributeFiles(files_list, nodes_list, peer, local_addr)
    } else {
        // torrent map already exists
        torrent_map = initTorrentMap()
    }

    // wait for incoming CNTF-Packets
    peer.SetTorrentMap(torrent_map)
    for true {
        peer.Run()
        saveTorrentMap(torrent_map)
    }
}
