package tornet

import (
    "encoding/json"
    "protocoldesign/pft"
    "io/ioutil"
    "path/filepath"
    "os"
)

func (this *Torrent) Write(parent_dir string) string {
    data, err := json.Marshal(this)
    pft.CheckError(err)

    file_path := filepath.Join(parent_dir, this.FilePath + ".torrent")
    os.MkdirAll(filepath.Dir(file_path), 0755)
    err = ioutil.WriteFile(file_path, data, 0755)
    pft.CheckError(err)
    return file_path
}

func (this *Torrent) Read(file_path string) {
    data, err := ioutil.ReadFile(file_path)
    pft.CheckError(err)

    err = json.Unmarshal(data, this)
    pft.CheckError(err)
}