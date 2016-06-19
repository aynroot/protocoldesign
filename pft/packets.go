package pft

import (
	"bytes"
	"encoding/binary"
	"errors"
	"golang.org/x/crypto/sha3"
	"fmt"
)

// calculates hash and concatenates packet to single []byte
func MakePacket(packet_type byte, payload []byte) []byte {
	var content []byte
	if payload == nil {
		content = []byte{packet_type}
	} else {
		content = append([]byte{packet_type}, payload...)
	}

	hash := sha3.Sum256(content)
	return append(hash[:16], content...)
}

func GetPacketType(packet []byte) byte {
	return packet[16]
}

func VerifyPacket(packet []byte, size int) bool {
	hash := sha3.Sum256(packet[16:size])
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

func EncodeReq(rid string) []byte {
	return MakePacket(REQ, []byte(rid))
}

func DecodeReq(packet []byte, size int) (error, string) {
	if size <= 17 { // 17 is length of hash + type
		return errors.New("packet too short"), ""
	}
	return nil, string(packet[17:size])
}

func EncodeData(block_index uint32, data_block []byte) []byte {
	return MakePacket(DATA, append(ToBigEndian32(block_index), data_block...))
}

func DecodeData(packet []byte, size int) (error, uint32, []byte) {
	if size <= 21 { // 17 byte header, 4 byte block index
		return errors.New("packet too short"), 0, nil
	}
	return nil, binary.BigEndian.Uint32(packet[17:21]), packet[21:size]
}

func EncodeReqNack() []byte {
	return MakePacket(REQ_NACK, nil)
}

func EncodePushAck() []byte {
	return MakePacket(PUSH_ACK, nil)
}

func EncodeRst() []byte {
	return MakePacket(RST, nil)
}

func EncodePush(rid string) []byte {
	return MakePacket(PUSH, []byte(rid))
}

func DecodePush(packet []byte, size int) (error, string) {
	if size <= 17 {
		return errors.New("packet too short"), ""
	}
	return nil, string(packet[17:size])
}

func EncodeGet(blockIndex uint32) []byte {
	return MakePacket(GET, ToBigEndian32(blockIndex))
}

func DecodeGet(packet []byte, size int) (error, uint32) {
	if size != 21 { // 17 byte header, 4 byte block index
		return errors.New("packet too short"), 0
	}
	return nil, binary.BigEndian.Uint32(packet[17:size])
}

func EncodeReqAck(resource_size uint64, resource_hash []byte) []byte {
	return MakePacket(REQ_ACK, append(ToBigEndian64(resource_size), resource_hash...))
}

func DecodeReqAck(packet []byte, size int) (error, uint64, []byte) {
	if size != 17 + 8 + 32 { // 17 byte header, 8 byte resource size, 32 byte resource hash
		return errors.New("invalid packet length"), 0, nil
	}

	resource_size := binary.BigEndian.Uint64(packet[17:25])
	return nil, resource_size, packet[25:size]
}

func EncodeCntf(chunk_rid string, info_byte byte) []byte {
	return MakePacket(CNTF, append([]byte{info_byte}, []byte(chunk_rid)...))
}

func DecodeCntf(packet []byte, size int) (error, byte, string) {
	if size <= 18 { // 17 byte header, 1 byte info byte
		return errors.New("packet too short"), 0, ""
	}
	fmt.Println(string(packet[22:size]))
	return nil, packet[17], string(packet[18:size])
}

