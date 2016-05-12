package pft

import (
    //"errors"
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
    F3    [495]byte     // RID
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
    F3    [4]byte       // Block index
    F4    [491]byte     // Data
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

    h := sha3.Sum256(tmp)

    //truncating
    for i := 0; i < 16; i++ {
        hash[i] = h[i]
    }

    return hash
}

func ResourceHash (p_type [1]byte, resource_size [8]byte) [32]byte {

    var res_hash  [32]byte
    var tmp   []byte

    // Specific field first + Packet_type
    tmp = append(resource_size[:])
    tmp = append(tmp, p_type[0])

    h := sha3.Sum256(tmp)

    //truncating
    for i := 0; i < 32; i++ {
        res_hash[i] = h[i]
    }

    return res_hash
}


func EncodeReq (rid string) (error, []byte) {

    var rid_bytes   [495]byte
    packet_type := [1]byte{1}
    temp := make([]byte, 495)

    if (len(rid) < 495) {
        // Getting the characters to the fixed size binary array
        for i := 0; i < len(rid); i++ {
            rid_bytes[i] = rid[i]
        }
        temp = rid_bytes[:]
        hash := GetSha3(packet_type,  temp)

        encode := Req{hash, packet_type, rid_bytes}
        buf := new(bytes.Buffer)
        err := binary.Write(buf, binary.BigEndian, &encode)

        return err, buf.Bytes()
    }

    return nil, nil // If the sanity check is not passed
}

func DecodeReq(packet []byte) (error, string, [1]byte) {

    var decoded Req
    err := binary.Read(bytes.NewReader(packet), binary.BigEndian, &decoded)

    result := string(bytes.Trim(decoded.F3[:495], "\x00"))
    return err, result, decoded.F2

}


func EncodeNack () (error, []byte) {

    packet_type := [1]byte{3}
    hash := GetSha3(packet_type, []byte{3})

    encode := Req_nack{hash, packet_type}
    buf := new(bytes.Buffer)
    err := binary.Write(buf, binary.BigEndian, &encode)

    return err, buf.Bytes()

}

func DecodeNack(packet []byte) (error, [1]byte) {

    var decoded Req_nack
    err := binary.Read(bytes.NewReader(packet), binary.BigEndian, &decoded)

    return err, decoded.F2
}

func EncodePushAck () (error, []byte) {

    packet_type := [1]byte{5}
    hash := GetSha3(packet_type, []byte{5})

    encode := Push_ack{hash, packet_type}
    buf := new(bytes.Buffer)
    err := binary.Write(buf, binary.BigEndian, &encode)

    return err, buf.Bytes()

}

func DecodePushAck(packet []byte) (error, [1]byte) {

    var decoded Push_ack
    err := binary.Read(bytes.NewReader(packet), binary.BigEndian, &decoded)

    return err, decoded.F2
}

func EncodeRst () (error, []byte) {

    packet_type := [1]byte{8}
    hash := GetSha3(packet_type, []byte{8})

    encode := Rst{hash, packet_type}
    buf := new(bytes.Buffer)
    err := binary.Write(buf, binary.BigEndian, &encode)

    return err, buf.Bytes()

}

func DecodeRst(packet []byte) (error, [1]byte) {

    var decoded Rst
    err := binary.Read(bytes.NewReader(packet), binary.BigEndian, &decoded)

    return err, decoded.F2
}

func EncodePush (rid string) (error, []byte) {

    var rid_bytes   [495]byte
    packet_type := [1]byte{4}
    temp := make([]byte, 495)

    if (len(rid) < 495) {
        // Getting the characters to the fixed size binary array
        for i := 0; i < len(rid); i++ {
            rid_bytes[i] = rid[i]
        }
        temp = rid_bytes[:]
        hash := GetSha3(packet_type,  temp)

        encode := Push{hash, packet_type, rid_bytes}
        buf := new(bytes.Buffer)
        err := binary.Write(buf, binary.BigEndian, &encode)

        return err, buf.Bytes()
    }

    return nil, nil // If the sanity check is not passed
}

func DecodePush(packet []byte) (error, string, [1]byte) {

    var decoded Push
    err := binary.Read(bytes.NewReader(packet), binary.BigEndian, &decoded)

    result := string(bytes.Trim(decoded.F3[:495], "\x00"))
    return err, result, decoded.F2

}


func EncodeGet (blockIndex [4]byte) (error, []byte) {

    packet_type := [1]byte{6}
    temp := make([]byte, 4)

    temp = blockIndex[:]
    hash := GetSha3(packet_type,  temp)

    encode := Get{hash, packet_type, blockIndex}
    buf := new(bytes.Buffer)
    err := binary.Write(buf, binary.BigEndian, &encode)

    return err, buf.Bytes()

}

func DecodeGet(packet []byte) (error, [4]byte, [1]byte) {

    var decoded Get
    err := binary.Read(bytes.NewReader(packet), binary.BigEndian, &decoded)

    return err, decoded.F3, decoded.F2

}


func EncodeData (block_index [4]byte, data_block [491]byte) (error, []byte) {

    packet_type := [1]byte{7}
    temp := make([]byte, 491)

    temp = data_block[:]
    hash := GetSha3(packet_type,  temp)

    encode := Data{hash, packet_type, block_index, data_block}
    buf := new(bytes.Buffer)
    err := binary.Write(buf, binary.BigEndian, &encode)

    return err, buf.Bytes()
}

func DecodeData(packet []byte) (error, [491]byte, [4]byte, [1]byte) {

    var decoded Data
    err := binary.Read(bytes.NewReader(packet), binary.BigEndian, &decoded)

    return err, decoded.F4, decoded.F3, decoded.F2

}


func EncodeReqAck (resource_size [8]byte) (error, []byte) {

    packet_type := [1]byte{2}
    temp := make([]byte, 8)

    temp = resource_size[:]
    hash := GetSha3(packet_type,  temp)

    resource_hash := ResourceHash(packet_type, resource_size)

    encode := Req_ack{hash, packet_type, resource_size, resource_hash}
    buf := new(bytes.Buffer)
    err := binary.Write(buf, binary.BigEndian, &encode)

    return err, buf.Bytes()
}

func DecodeReqAck(packet []byte) (error, [32]byte, [8]byte, [1]byte) {

    var decoded Req_ack
    err := binary.Read(bytes.NewReader(packet), binary.BigEndian, &decoded)

    return err, decoded.F4, decoded.F3, decoded.F2

}
