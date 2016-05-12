package pft

import (
    "errors"
    "encoding/binary"
    "bytes"
    "crypto/sha3"
)


type Req struct {
    F1    [16]byte      // SHA3
    F2    [1]byte       // Type
    F3    [495]byte     // RID
}

type Req_ack struct {
    F1    [16]byte      // SHA3
    F2    [1]byte       // Type
    F3    [8]byte       // Resource Size
    F4    [32]byte      // Resource Hash
}

type Req_nack struct {
    F1    [16]byte      // SHA3
    F2    [1]byte       // Type
}

type Push struct {
    F1    [16]byte      // SHA3
    F2    [1]byte       // Type
}

type Push_ack struct {
    F1    [16]byte      // SHA3
    F2    [1]byte       // Type
}

type Get struct {
    F1    [16]byte      // SHA3
    F2    [1]byte       // Type
    F3    [4]byte       // Block index
}

type Data struct {
    F1    [16]byte      // SHA3
    F2    [1]byte       // Type
}

type Rst struct {
    F1    [16]byte      // SHA3
    F2    [1]byte       // Type
}

func GetSha3 (p_type [1]byte, field []byte) [16]byte {

    var hash  [16]byte
    var tmp   []byte

    // Specific field first + Packet_type
    tmp = append(field[:])
    tmp = append(tmp, p_type[0])

    h := make([]byte, 64)

    sha3.ShakeSum256(h, tmp)

    //truncating
    for i := 0; i < 16; i++ {
        hash[i] = h[i]
    }

    return hash
}

func EncodeReq (rid string) (error, []byte) {

    var rid_bytes   [495]byte
    var t1 Req
    packet_type := [1]byte{1}
    temp := make([]byte, 495)

    if (len(rid) < 495) {
        // Getting the characters to the fixed size binary array
        for i := 0; i < len(rid); i++ {
            rid_bytes[i] = rid[i]
        }
        temp = rid_bytes[:]
        hash := GetSha3(packet_type,  temp)

        t1 = Req{hash, packet_type, rid_bytes}
        buf := new(bytes.Buffer)
        err := binary.Write(buf, binary.BigEndian, &t1)

        return err, buf.Bytes()
    }

    return nil, nil // If the sanity check is not passed
}

func DecodeReq(packet []byte) (error, string) {

    var decoded Req
    err := binary.Read(bytes.NewReader(packet), binary.BigEndian, &decoded)

    result := string(bytes.Trim(decoded.F3[:495], "\x00"))
    return err, result

}

func EncodeReqAck() [] byte {
    return nil
}

func DecodeReqAck() error {
    return errors.New("Not implemented")
}


// ... to be continued