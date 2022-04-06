package applications

import (
	"log"
	"net"
	"scheduleupdater-server/addrtranslation"
)

type ApplicationHelloWorld struct{}

func NewApplicationHelloWorld() ApplicationHelloWorld {
	return ApplicationHelloWorld{}
}

func (app ApplicationHelloWorld) Type() AppType {
	return AppTypeHelloWorld
}

func (app ApplicationHelloWorld) ProcessPacket(addr *net.UDPAddr, packet []byte) {
	addrIP := addrtranslation.AddrToIPString(addr)
	log.Println("Received an hello world packet from:", addrIP, " with content:", packet[0])
}
