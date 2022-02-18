package udpack

import (
	"errors"
	"fmt"
)

type PacketType uint8

const (
	PacketTypeData = iota
	PacketTypeAck
)

func newAckPacket(sequenceNumber uint8) ([]byte, error) {
	header, err := newHeader(PacketTypeAck, sequenceNumber)
	if err != nil {
		return nil, err
	}
	return []byte{byte(header)}, nil
}

func newDataPacket(sequenceNumber uint8, packet []byte) ([]byte, error) {
	header, err := newHeader(PacketTypeData, sequenceNumber)
	if err != nil {
		return nil, err
	}
	packetWithHeader := make([]byte, len(packet)+1)
	packetWithHeader[0] = byte(header)
	copy(packetWithHeader[1:], packet)
	return packetWithHeader, nil
}

type Header uint8

func newHeader(packetType PacketType, sequenceNumber uint8) (Header, error) {
	header := Header(0)
	header.encodePacketType(packetType)
	err := header.encodeSequenceNumber(sequenceNumber)
	if err != nil {
		return 0, err
	}
	return header, nil
}

func (header *Header) encodePacketType(packetType PacketType) {
	const packetTypeShift uint8 = 7
	if packetType == PacketTypeData {
		*header |= Header(uint8(PacketTypeData) << packetTypeShift)
	} else if packetType == PacketTypeAck {
		*header |= Header(uint8(PacketTypeAck) << packetTypeShift)
	}
}

func DecodePacketType(header Header) PacketType {
	const packetTypeMask uint8 = 0x80
	isAck := header&Header(packetTypeMask) != 0
	if isAck {
		return PacketTypeAck
	}
	return PacketTypeData
}

func (header *Header) encodeSequenceNumber(sequenceNumber uint8) error {
	const maxSequenceNumber = 0b01111111
	if sequenceNumber > maxSequenceNumber {
		return errors.New(fmt.Sprintf(
			"The sequence number given %d is greater than the maximum sequence number authorized which is %d",
			sequenceNumber,
			maxSequenceNumber))
	}
	*header |= Header(sequenceNumber)
	return nil
}

func decodeSequenceNumber(header Header) uint8 {
	const sequenceNumberMask = 0x7F
	return uint8(header) & sequenceNumberMask
}

// RemoveHeaderFromPacket remove the header that contains a PacketType and a SequenceNumber
// The header is a byte that has the following structure:
// 0b1 0000111
//   ^ ^^^^^^^
//   | |
//   | +------------ Sequence Number (last 7 bits)
//   +-- PacketType (first bit)
func RemoveHeaderFromPacket(packet []byte) (Header, []byte) {
	return Header(packet[0]), packet[1:]
}
