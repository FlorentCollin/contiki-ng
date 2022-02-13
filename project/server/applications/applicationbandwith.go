package applications

import (
	"coap-server/udpack"
	"errors"
	"fmt"
	"log"
	"net"
)

type ApplicationBandwith struct {
	bandwith map[udpack.IPString]uint
}

func NewApplicationBandwith() ApplicationBandwith {
	return ApplicationBandwith{
		bandwith: make(map[udpack.IPString]uint),
	}
}

func (app ApplicationBandwith) Type() AppType {
	return AppTypeBandwith
}

func (app ApplicationBandwith) ProcessPacket(addr *net.UDPAddr, packet []byte) {
	addrIP := udpack.AddrToIPString(addr)
	bandwith, err := decodeBandwith(packet)
	if err != nil {
		log.Panic(err)
	}
	app.bandwith[addrIP] = bandwith
}

func decodeBandwith(packet []byte) (uint, error) {
	if len(packet) < 1 || len(packet) > 1 {
		return 0, errors.New(fmt.Sprintf(
			"The bandwith packet in malformated, its length must be 1 byte and the packet received is %d bytes long.",
			len(packet)))
	}
	return uint(packet[0]), nil
}
