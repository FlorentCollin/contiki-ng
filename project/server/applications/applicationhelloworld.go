package applications

import (
	"log"
	"net"
	"scheduleupdater-server/addrtranslation"
)

// ApplicationHelloWorld simple example of an application that prints out
// the content of a packet when one is received.
type ApplicationHelloWorld struct{}

func NewApplicationHelloWorld() ApplicationHelloWorld {
	return ApplicationHelloWorld{}
}

func (app ApplicationHelloWorld) Type() AppType {
	return AppTypeHelloWorld
}

func (app ApplicationHelloWorld) ProcessPacket(addr *net.UDPAddr, packet []byte) {
	addrIP := addrtranslation.AddrToIPString(addr)
	log.Printf("Received an hello world packet from: %s with content: %+v", addrIP, packet)
}
