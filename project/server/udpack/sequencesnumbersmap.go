package udpack

import "net"

type IPString string

func AddrToIPString(addr *net.UDPAddr) IPString {
	return IPString(addr.IP.String())
}

type SequencesNumbersMap map[IPString]uint8

func (sequenceNumberMap SequencesNumbersMap) initializeAddrIP(addrIP IPString) uint8 {
	initialSequenceNumber := uint8(1)
	sequenceNumberMap[addrIP] = initialSequenceNumber
	return initialSequenceNumber
}

func (sequenceNumberMap SequencesNumbersMap) increment(addrIP IPString, currentSequenceNumber uint8) {
	if currentSequenceNumber > maxSequenceNumber {
		sequenceNumberMap[addrIP] = 0
		return
	}
	sequenceNumberMap[addrIP] += 1
}

func (sequenceNumberMap SequencesNumbersMap) expected(addrIP IPString) uint8 {
	if expectedSequenceNumber, in := sequenceNumberMap[addrIP]; in {
		return expectedSequenceNumber
	}
	return sequenceNumberMap.initializeAddrIP(addrIP)
}

const maxSequenceNumber = 1 << 7
