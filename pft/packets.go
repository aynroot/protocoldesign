package pft

import (
    //"errors"
    "encoding/binary"
    "bytes"
    "golang.org/x/crypto/sha3"
    "errors"
)

// calculates hash and concatenates packet to single []byte
func MakePacket(packet_type byte, payload []byte) []byte {
    var content []byte
    if(payload == nil) {
        content = []byte{packet_type}
    } else {
        content = append([]byte{packet_type}, payload...)
    }

    hash := sha3.Sum256(content)
    return append(hash[:16], content...)
}


func VerifyPacket(packet []byte) bool {
    hash := sha3.Sum256(packet[16:])
    return bytes.Equal(packet[:16], hash[:16])
}

func ToBigEndian32(num uint32) []byte {
    num_big_endian := make([]byte, 4)
    binary.BigEndian.PutUint32(num_big_endian, num)
    return num_big_endian
}

func ToBigEndian64(num uint64) []byte {
    num_big_endian := make([]byte, 8)
    binary.BigEndian.PutUint64(num_big_endian, num)
    return num_big_endian
}

func EncodeReq (rid string) []byte {
    return MakePacket(1, []byte(rid))
}

func DecodeReq(packet []byte) (error, string) {
    if len(packet) <= 17 { // 17 is length of hash + type
        return errors.New("packet too short"), ""
    }

    return nil, string(packet[17:])
}

func EncodeData (block_index uint32, data_block []byte) []byte {
    return MakePacket(7, append(ToBigEndian32(block_index), data_block...))
}

func DecodeData(packet []byte) (error, uint32, []byte) {
    if len(packet) <= 21 { // 21 is length of hash + type + data_block
        return errors.New("packet too short"), 0, nil
    }

    return nil, binary.BigEndian.Uint32(packet[17:21]), packet[21:]
}

func EncodeNack () []byte {
    return MakePacket(3, nil)
}

func EncodePushAck() []byte {
    return MakePacket(5, nil)

}

func EncodeRst() []byte {
    return MakePacket(8, nil)
}


func EncodePush (rid string) []byte {
    return MakePacket(4, []byte(rid))
}

func DecodePush(packet []byte) (error, string) {
    if len(packet) <= 17 {
        return errors.New("packet too short"), ""
    }

    return nil, string(packet[17:])

}


func EncodeGet (blockIndex uint32) []byte {
    return MakePacket(6, ToBigEndian32(blockIndex))

}

func DecodeGet(packet []byte) (error, uint32) {
    if len(packet) != 21 {
        return errors.New("packet too short"), 0
    }

    return nil, binary.BigEndian.Uint32(packet[17:21])
}


func EncodeReqAck (resource_size uint64, resource_hash []byte) []byte {
    return MakePacket(2, append(ToBigEndian64(resource_size), resource_hash...))
}

func DecodeReqAck(packet []byte) (error, uint64, []byte) {
    if len(packet) != 17 + 8 + 32 { // 17 byte header, 8 byte resource size, 32 byte resource hash
        return errors.New("invalid packet length"), 0, nil
    }

    resource_size := binary.BigEndian.Uint64(packet[17:25])

    return nil, resource_size, packet[25:57]
}