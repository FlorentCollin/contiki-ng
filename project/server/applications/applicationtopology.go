package applications

import (
	"coap-server/udpack"
	"errors"
	"fmt"
	"log"
	"net"
)

type ApplicationTopology struct {
	Topology Topology
	nClient  uint
}

func NewApplicationTopology(nClients uint) ApplicationTopology {
	return ApplicationTopology{
		Topology: newTopology(),
		nClient:  nClients,
	}
}

func (app *ApplicationTopology) Type() AppType {
	return AppTypeTopology
}

func (app *ApplicationTopology) ProcessPacket(addr *net.UDPAddr, packet []byte) {
	log.Println("Received a topology packet")
	addrIP := udpack.AddrToIPString(addr)
	topologyPacket, err := decodeTopologyPacket(packet)
	if err != nil {
		log.Panic(err)
	}
	app.Topology.SetNeighbors(addrIP, topologyPacket.Neighbors)
}

func (app *ApplicationTopology) Ready() bool {
	return len(app.Topology) == int(app.nClient)
}
func decodeTopologyPacket(packet []byte) (*TopologyPacket, error) {
	const ipSize = 16
	if len(packet)%ipSize != 0 {
		return nil, errors.New(
			fmt.Sprintf("the length of a Topology packet should be a multiple of %d"+
				" since each IP is represented with 16 bytes, currently the length is %d", ipSize, len(packet)))
	}
	neighborsCount := len(packet) / ipSize
	neighbors := make([]udpack.IPString, 0, neighborsCount)
	for i := 0; i < neighborsCount; i++ {
		neighborIP := packet[i*ipSize : i*ipSize+ipSize]
		neighborIPString := udpack.NetIPToIPString(neighborIP).LinkLocalToGlobal()
		neighbors = append(neighbors, neighborIPString)
	}

	return &TopologyPacket{Neighbors: neighbors}, nil
}

type TopologyPacket struct {
	Neighbors []udpack.IPString
}

type Topology map[udpack.IPString][]udpack.IPString

func newTopology() Topology {
	return make(Topology)
}

func (topology Topology) AddNeighbor(addrIP, neighborIP udpack.IPString) {
	topology[addrIP] = append(topology[addrIP], neighborIP)
}

func (topology Topology) SetNeighbors(addrIP udpack.IPString, neighborsIP []udpack.IPString) {
	log.Println("Settings new neighbors")
	topology[addrIP] = neighborsIP
}

func (topology Topology) ClearNeighbors(addrIP udpack.IPString) {
	topology[addrIP] = []udpack.IPString{}
}
