package applications

import (
	"coap-server/udpack"
	"errors"
	"fmt"
	"log"
	"net"
)

type BandwidthMap map[udpack.IPString]uint
type ApplicationBandwidth struct {
	Bandwith BandwidthMap
	nClients uint
}

func NewApplicationBandwidth(nClients uint) ApplicationBandwidth {
	return ApplicationBandwidth{
		Bandwith: make(BandwidthMap),
		nClients: nClients,
	}
}

func (app *ApplicationBandwidth) Type() AppType {
	return AppTypeBandwidth
}

func (app *ApplicationBandwidth) ProcessPacket(addr *net.UDPAddr, packet []byte) {
	addrIP := udpack.AddrToIPString(addr)
	bandwith, err := decodeBandwith(packet)
	if err != nil {
		log.Panic(err)
	}
	app.Bandwith[addrIP] = bandwith
}

func (app *ApplicationBandwidth) Ready() bool {
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
