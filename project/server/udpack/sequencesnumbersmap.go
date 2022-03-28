package udpack

import (
	"net"
	"strings"
)

type IPString string

func AddrToIPString(addr *net.UDPAddr) IPString {
	return IPString(addr.IP.String())
}

func NetIPToIPString(addr net.IP) IPString {
	return IPString(addr.String())
}

func (ipString IPString) GlobalToLinkLocal() IPString {
    // TODO: Normally this should not be done, we need a better way to translate link local to global addresses.
    return IPString(strings.Replace(string(ipString), "fd00", "fe80", 1))
}

func (ipString IPString) LinkLocalToGlobal() IPString {
	// TODO: This is a kind of hack since normally we should not be able to get the local address from the link-local addr
	// This certainly only works in Cooja
	return IPString(strings.Replace(string(ipString), "fe80", "fd00", 1))
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
