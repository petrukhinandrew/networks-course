package main

import (
	"checksum"
	"errors"
)

type Packet struct {
	index    uint8
	length   uint8
	checksum uint16
	data     []byte
}

var (
	ErrBadPacketSize     = errors.New("bad packet size")
	ErrBadPacketLength   = errors.New("bad packet length")
	ErrBadPacketCheckSum = errors.New("bad packet checksum")
)

func (p *Packet) Bytes() []byte {
	return append([]byte{p.index, p.length, byte(p.checksum >> 8), byte(p.checksum)}, p.data...)
}
func AckPacket(index int) *Packet {
	data := []byte("ACK")
	return &Packet{index: uint8(index), length: uint8(4 + len(data)), checksum: checksum.CheckSum(append([]byte{uint8(index), uint8(4 + len(data))}, data...)), data: data}
}
func EncodePacket(index int, data []byte) *Packet {
	p := &Packet{index: uint8(index), length: uint8(4 + len(data)), checksum: checksum.CheckSum(append([]byte{uint8(index), uint8(4 + len(data))}, data...)), data: data}
	return p
}

func DecodePacket(src []byte) (*Packet, error) {
	if len(src) < 4 {
		return nil, ErrBadPacketSize
	}
	np := &Packet{}
	np.index = src[0]
	np.length = src[1]
	np.checksum = uint16(src[2])<<8 + uint16(src[3])
	np.data = src[4:]
	if !checksum.ValidatePacket(np.Bytes()) {
		return nil, ErrBadPacketCheckSum
	}
	if len(np.data)+4 != int(np.length) {
		return nil, ErrBadPacketLength
	}
	return np, nil
}
