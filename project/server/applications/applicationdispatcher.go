package applications

import (
	"log"
	"net"
)

type App interface {
	ProcessPacket(addr *net.UDPAddr, packet []byte)
	Type() AppType
}

type AppType int

const (
	AppTypeGraph = iota
	AppTypeTopology
	AppTypeBandwith
	AppTypeHelloWorld
	ApplicationTypeAll
)

func (appType AppType) debug() string {
	return [...]string{
		"AppTypeGraph",
		"AppTypeTopology",
		"AppTypeBandwith",
		"AppTypeHelloWorld",
		"ApplicationTypeAll",
	}[appType]
}

type AppDispatcher struct {
	applications [ApplicationTypeAll][]App
}

func NewAppDispatcher() *AppDispatcher {
	return &AppDispatcher{}
}

func (dispatcher *AppDispatcher) Subscribe(app App) *AppDispatcher {
	appType := app.Type()
	dispatcher.applications[appType] = append(dispatcher.applications[appType], app)
	return dispatcher
}

func (dispatcher *AppDispatcher) Handler(addr *net.UDPAddr, packet []byte) {
	log.Println("Application dispatcher handler called")
	appType, packetWithoutAppType := removeAppType(packet)
	if appType >= ApplicationTypeAll {
		log.Panic("The AppType contained in the packet is not a valid AppType (value: ", appType, ")")
	}
	log.Println("Apptype:", appType.debug())
	for _, app := range dispatcher.applications[appType] {
		go app.ProcessPacket(addr, packetWithoutAppType)
	}
}

func removeAppType(packet []byte) (AppType, []byte) {
	rawAppType := packet[0]
	return AppType(rawAppType), packet[1:]
}
