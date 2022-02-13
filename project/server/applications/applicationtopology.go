package applications

import (
	"coap-server/udpack"
	"log"
	"net"
)

type ApplicationTopology struct {
	topology map[udpack.IPString][]udpack.IPString
}

func NewApplicationTopology() ApplicationTopology {
	return ApplicationTopology{
		topology: newTopology(),
	}
}

func (app ApplicationTopology) Type() AppType {
	return AppTypeTopology
}

func (app ApplicationTopology) ProcessPacket(addr *net.UDPAddr, packet []byte) {
	// TODO Parse the packet correctly and update the topology graph
	// but before that the packet representation should be figured out
	log.Println("Received a topology packet")
	//addrIP := udpack.AddrToIPString(addr)
	//topologyPacket, err := decodeTopologyPacket(packet)
	//if err != nil {
	//	log.Panic(err)
	//}
}

type Topology map[udpack.IPString][]udpack.IPString

func newTopology() Topology {
	return make(Topology)
}

func (topology Topology) addNeighbor(addrIP, neighborIP udpack.IPString) {
	topology[addrIP] = append(topology[addrIP], neighborIP)
}

func (topology Topology) clearNeighbors(addrIP udpack.IPString) {
	topology[addrIP] = []udpack.IPString{}
}
