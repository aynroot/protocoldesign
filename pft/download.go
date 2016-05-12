package pft


// this packet deals with resource downloads
// a file format for storing partial downloads needs to be developed
// a partial download file needs to save the rid, the size, the hash and the already downloaded data of the resource

type download struct {
    rid string
    size uint64
    hash []byte
    received_index uint32
    requested_index uint32
    local_file_path string
}


// if index == received_index + 1: writes the payload to disk and updates received_index
// else: set requested_index = received_index
// returns true if the payload was written to disk
func (this download) HandleDataPacket(index uint32, payload []byte) bool {
    return false
}

// loads the internal variables from a partial download file
func (this download) LoadPartialDownload(file_path string) {

}

func (this download) IsFinished() bool {
    //return (this.received_index + 1) * 491 >= this.size
    return false
}

// creates get packet for requested_index + 1 (use encode function from packets.go)
func (this download) CreateNextGet() []byte {
    return nil
}