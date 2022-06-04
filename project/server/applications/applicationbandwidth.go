package applications

import (
	"errors"
	"fmt"
	"log"
	"net"
	"scheduleupdater-server/addrtranslation"
	"sync"
)

type BandwidthMap map[addrtranslation.IPString]uint

// ApplicationBandwidth gathers all the bandwidth requirements of the network's nodes.
// The bandwidth requirement is an integer indicating the number of packets that the node
// wants to send in a single slotframe.
type ApplicationBandwidth struct {
	Bandwith BandwidthMap
	nClients uint
	lock     sync.RWMutex
}

func NewApplicationBandwidth(nClients uint) ApplicationBandwidth {
	return ApplicationBandwidth{
		Bandwith: make(BandwidthMap),
		nClients: nClients,
		lock:     sync.RWMutex{},
	}
}

func (app *ApplicationBandwidth) Type() AppType {
	return AppTypeBandwidth
}

func (app *ApplicationBandwidth) ProcessPacket(addr *net.UDPAddr, packet []byte) {
	addrIP := addrtranslation.AddrToIPString(addr)
	bandwith, err := decodeBandwith(packet)
	if err != nil {
		log.Panic(err)
	}
	app.lock.Lock()
	defer app.lock.Unlock()
	app.Bandwith[addrIP] = bandwith
}

func (app *ApplicationBandwidth) Ready() bool {
	return len(app.Bandwith) == int(app.nClients)
}

func decodeBandwith(packet []byte) (uint, error) {
	if len(packet) < 1 {
		return 0, errors.New(fmt.Sprintf(
			"The bandwith packet in malformated, its length must be 1 byte and the packet received is %d bytes long.",
			len(packet)))
	}
	return uint(packet[0]), nil
}
