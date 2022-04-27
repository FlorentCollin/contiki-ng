package applications

import (
	"log"
	"net"
)

// App an application allows clients to communicate with the server
// by sending packet identified by an AppType unique to the application.
// Each App must have a method named `ProcessPacket` that handles each
// packet received from a client.
type App interface {
	ProcessPacket(addr *net.UDPAddr, packet []byte)
	Type() AppType
}

type AppType int

const (
	AppTypeGraph = iota
	AppTypeTopology
	AppTypeBandwidth
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

// AppDispatcher dispatches the packets by their AppType. AppDispatcher allows
// applications to subscribe to the AppDispatcher to receive the packets with
// their AppType.
type AppDispatcher struct {
	applications [ApplicationTypeAll][]App
}

func NewAppDispatcher() *AppDispatcher {
	return &AppDispatcher{}
}

// Subscribe subscribes an application to the AppDispatcher to receive
// packets with their AppType.
func (dispatcher *AppDispatcher) Subscribe(app App) *AppDispatcher {
	appType := app.Type()
	dispatcher.applications[appType] = append(dispatcher.applications[appType], app)
	return dispatcher
}

// Handler handles a `packet` coming from an address `addr` and dispatches
// the `packet` to the corresponding applications.
func (dispatcher *AppDispatcher) Handler(addr *net.UDPAddr, packet []byte) {
	appType, packetWithoutAppType := removeAppType(packet)
	if appType >= ApplicationTypeAll {
		log.Panic("The AppType contained in the packet is not a valid AppType"+
			" (value: ", appType, ")")
	}
	for _, app := range dispatcher.applications[appType] {
		go app.ProcessPacket(addr, packetWithoutAppType)
	}
}

func removeAppType(packet []byte) (AppType, []byte) {
	rawAppType := packet[0]
	return AppType(rawAppType), packet[1:]
}
