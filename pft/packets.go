package pft

import "errors"

// encode and decode function for each packet type
// the encode functions take the respective packets' arguments and return a finished packet (including hash and type header)
// the decode functions return the packets arguments from a complete packet byte slice.
// NOTE: the first return argument of each encode/decode function MUST be an error.
// On success, the functions must return nil as the error return value.
// On failure, they must return a suitable error description

func EncodeReq (rid string) []byte {
    return nil
}

func DecodeReq(packet []byte) (error, string) {
    return errors.New("Not implemented"), ""
}

func EncodeReqAck() [] byte {
    return nil
}

func DecodeReqAck() (error, string) {
    return errors.New("Not implemented"), ""
}



// ... to be continued