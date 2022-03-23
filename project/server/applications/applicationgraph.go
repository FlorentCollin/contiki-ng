package applications

import (
	"coap-server/udpack"
	"encoding/binary"
	"fmt"
	"log"
	"net"
	"time"
    "sync"
)

type RPLGraph map[udpack.IPString]*RPLLink

type ApplicationGraph struct {
	Graph    RPLGraph
	nClients uint
    lock sync.RWMutex

}

func NewApplicationGraph(nClients uint) ApplicationGraph {
	return ApplicationGraph{
		Graph:    make(RPLGraph),
		nClients: nClients,
        lock: sync.RWMutex{},
	}
}

func (app *ApplicationGraph) Type() AppType {
	return AppTypeGraph
}

func (app *ApplicationGraph) ProcessPacket(_ *net.UDPAddr, packet []byte) {
	log.Println("Processing graph packet...")
	graphUpdates := decodeGraphUpdateData(packet)
	for _, update := range graphUpdates {
		app.updateGraph(&update)
	}
}

func (app *ApplicationGraph) Ready() bool {
	log.Println("Current GRAPH SIZE: ", len(app.Graph))
	if len(app.Graph) > int(app.nClients-1) {
		fmt.Println("Something wrong...")
	}
	return len(app.Graph) == int(app.nClients-1) // -1 since we count the number of links
}

func (app *ApplicationGraph) updateGraph(graphUpdate *GraphTopologyUpdate) {
	childIPString := udpack.IPString(graphUpdate.ChildIP.String())
	parentIPString := udpack.IPString(graphUpdate.ParentIP.String())
	if v, in := app.Graph[childIPString]; !in || v.ParentIP != parentIPString {
		log.Printf("Adding RPL Link: from %s to %s\n", childIPString, parentIPString)
		app.Graph[childIPString] = &RPLLink{
			ParentIP:    parentIPString,
			ExpiredTime: time.Now().Add(time.Duration(graphUpdate.Lifetime) * time.Second),
		}
		go app.removeRPLLinkWhenLifetimeExpired(childIPString, graphUpdate.Lifetime)
	}
}

func (app *ApplicationGraph) removeRPLLinkWhenLifetimeExpired(childIP udpack.IPString, lifetime uint32) {
	time.Sleep(time.Duration(lifetime) * time.Second)
    app.lock.RLock()
    defer app.lock.RUnlock()
	rplLink, in := app.Graph[childIP]
	if !in || time.Now().Before(rplLink.ExpiredTime) {
		return
	}

	// Delete the child only if there is not another node that have a link to this child
	for _, link := range app.Graph {
		if link.ParentIP == childIP {
			return
		}
	}
	delete(app.Graph, childIP)
}

type RPLLink struct {
	ParentIP    udpack.IPString
	ExpiredTime time.Time
}

type GraphTopologyUpdate struct {
	ChildIP  net.IP
	ParentIP net.IP
	Lifetime uint32
}

func decodeGraphUpdateData(data []byte) []GraphTopologyUpdate {
	// TODO: refactor this part, and don't rely on magic number
	const updateSize = 16*2 + 4
	updates := make([]GraphTopologyUpdate, len(data)/updateSize)
	for i := 0; i < len(updates); i += 1 {
		// something is maybe wrong here since we get an addr fd00::2201:208:8:8:8
		updates[i] = GraphTopologyUpdate{
			ChildIP:  data[i*updateSize : i*updateSize+16],
			ParentIP: data[i*updateSize+16 : i*updateSize+32],
			Lifetime: binary.LittleEndian.Uint32(data[i*updateSize+32 : i*updateSize+32+4]),
		}
	}
	return updates
}
