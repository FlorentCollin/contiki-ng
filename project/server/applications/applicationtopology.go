package applications

import (
	"errors"
	"fmt"
	"log"
	"net"
	"scheduleupdater-server/addrtranslation"
	"sync"
)

// ApplicationTopology gathers the topology of the network. The topology, in this case,
// is the neighbors of the nodes.
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
	addrIP := addrtranslation.AddrToIPString(addr)
	topologyPacket, err := decodeTopologyPacket(packet)
	if err != nil {
		log.Panic(err)
	}
	app.Topology.SetNeighbors(addrIP, topologyPacket)
}

func (app *ApplicationTopology) Ready() bool {
	return len(app.Topology.TopologyMap) == int(app.nClient)
}

func decodeTopologyPacket(packet []byte) (*TopologyPacket, error) {
	const macSize = 8
	if len(packet)%8 != 0 {
		return nil, errors.New(
			fmt.Sprintf("the length of a Topology packet should be a multiple of %d"+
				" since each MAC addr is represented with %d bytes, currently the length is %d", macSize, macSize, len(packet)))
	}
	moteAddr := (*addrtranslation.MacAddr)(packet[0:macSize])
	neighborsCount := len(packet)/macSize - 1
	neighbors := make([]*addrtranslation.MacAddr, 0, neighborsCount)
	for i := 1; i <= neighborsCount; i++ {
		neighborMac := (*addrtranslation.MacAddr)(packet[i*macSize : i*macSize+macSize])
		neighbors = append(neighbors, neighborMac)
	}

	return &TopologyPacket{MoteAddr: moteAddr, Neighbors: neighbors}, nil
}

type TopologyPacket struct {
	MoteAddr  *addrtranslation.MacAddr
	Neighbors []*addrtranslation.MacAddr
}

type Topology struct {
	TopologyMap      map[addrtranslation.IPString][]*addrtranslation.MacAddr
	MacIPTranslation addrtranslation.MacIPTranslation
	lock             sync.RWMutex
}

func newTopology() Topology {
	return Topology{
		TopologyMap:      map[addrtranslation.IPString][]*addrtranslation.MacAddr{},
		MacIPTranslation: addrtranslation.NewMacIPTranslation(),
		lock:             sync.RWMutex{},
	}
}

func (topology *Topology) SetNeighbors(addrIP addrtranslation.IPString, packet *TopologyPacket) {
	topology.lock.Lock()
	defer topology.lock.Unlock()
	topology.TopologyMap[addrIP] = packet.Neighbors
	topology.MacIPTranslation.Add(packet.MoteAddr, addrIP)
}

func (topology *Topology) ClearNeighbors(addrIP addrtranslation.IPString) {
	topology.lock.Lock()
	defer topology.lock.Unlock()
	topology.TopologyMap[addrIP] = []*addrtranslation.MacAddr{}
}
