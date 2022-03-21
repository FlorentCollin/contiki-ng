package applications

import (
	"coap-server/udpack"
	"errors"
	"fmt"
	"log"
	"net"
)

type BandwidthMap map[udpack.IPString]uint
type ApplicationBandwith struct {
	Bandwith BandwidthMap
	nClients uint
}

func NewApplicationBandwith(nClients uint) ApplicationBandwith {
	return ApplicationBandwith{
		Bandwith: make(BandwidthMap),
		nClients: nClients,
	}
}

func (app *ApplicationBandwith) Type() AppType {
	return AppTypeBandwidth
}

func (app *ApplicationBandwith) ProcessPacket(addr *net.UDPAddr, packet []byte) {
	addrIP := udpack.AddrToIPString(addr)
	bandwith, err := decodeBandwith(packet)
	if err != nil {
		log.Panic(err)
	}
	app.Bandwith[addrIP] = bandwith
}

func (app *ApplicationBandwith) Ready() bool {
	return len(app.Bandwith) == int(app.nClients)
}

func decodeBandwith(packet []byte) (uint, error) {
	if len(packet) < 1 || len(packet) > 1 {
		return 0, errors.New(fmt.Sprintf(
			"The bandwith packet in malformated, its length must be 1 byte and the packet received is %d bytes long.",
			len(packet)))
	}
	return uint(packet[0]), nil
}
