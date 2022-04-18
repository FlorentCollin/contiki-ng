package udpack

import (
	"scheduleupdater-server/addrtranslation"
	"sync"
)

const maxSequenceNumber = 0b00111111

type SequenceNumbersMap struct {
	expectedSequenceNumbers map[addrtranslation.IPString]uint8
	lock                    sync.RWMutex
}

func NewSequenceNumbersMap() SequenceNumbersMap {
	return SequenceNumbersMap{
		expectedSequenceNumbers: make(map[addrtranslation.IPString]uint8),
		lock:                    sync.RWMutex{},
	}
}

func (sequenceNumberMap *SequenceNumbersMap) initializeAddrIP(addrIP addrtranslation.IPString) uint8 {
	sequenceNumberMap.lock.Lock()
	defer sequenceNumberMap.lock.Unlock()
	initialSequenceNumber := uint8(1)
	sequenceNumberMap.expectedSequenceNumbers[addrIP] = initialSequenceNumber
	return initialSequenceNumber
}

func (sequenceNumberMap *SequenceNumbersMap) increment(addrIP addrtranslation.IPString, currentSequenceNumber uint8) {
	sequenceNumberMap.lock.Lock()
	defer sequenceNumberMap.lock.Unlock()
	if currentSequenceNumber >= maxSequenceNumber {
		sequenceNumberMap.expectedSequenceNumbers[addrIP] = 0
		return
	}
	sequenceNumberMap.expectedSequenceNumbers[addrIP] += 1
}

func (sequenceNumberMap *SequenceNumbersMap) expected(addrIP addrtranslation.IPString) uint8 {
	sequenceNumberMap.lock.Lock()
	if expectedSequenceNumber, in := sequenceNumberMap.expectedSequenceNumbers[addrIP]; in {
		sequenceNumberMap.lock.Unlock()
		return expectedSequenceNumber
	}
	sequenceNumberMap.lock.Unlock()
	return sequenceNumberMap.initializeAddrIP(addrIP)
}
