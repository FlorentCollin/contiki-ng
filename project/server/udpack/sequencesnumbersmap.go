package udpack

import (
	"net"
	"strings"
	"sync"
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

const maxSequenceNumber = 0b00111111

type SequenceNumbersMap struct {
	expectedSequenceNumbers map[IPString]uint8
	lock                    sync.RWMutex
}

func NewSequenceNumbersMap() SequenceNumbersMap {
	return SequenceNumbersMap{
		expectedSequenceNumbers: make(map[IPString]uint8),
		lock:                    sync.RWMutex{},
	}
}

func (sequenceNumberMap *SequenceNumbersMap) initializeAddrIP(addrIP IPString) uint8 {
	initialSequenceNumber := uint8(1)
	sequenceNumberMap.expectedSequenceNumbers[addrIP] = initialSequenceNumber
	return initialSequenceNumber
}

func (sequenceNumberMap *SequenceNumbersMap) increment(addrIP IPString, currentSequenceNumber uint8) {
	sequenceNumberMap.lock.RLock()
	defer sequenceNumberMap.lock.RUnlock()
	if currentSequenceNumber > maxSequenceNumber {
		sequenceNumberMap.expectedSequenceNumbers[addrIP] = 0
		return
	}
	sequenceNumberMap.expectedSequenceNumbers[addrIP] += 1
}

func (sequenceNumberMap *SequenceNumbersMap) expected(addrIP IPString) uint8 {
	if expectedSequenceNumber, in := sequenceNumberMap.expectedSequenceNumbers[addrIP]; in {
		return expectedSequenceNumber
	}
	return sequenceNumberMap.initializeAddrIP(addrIP)
}
