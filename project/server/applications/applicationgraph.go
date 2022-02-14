package applications

import (
	"coap-server/udpack"
	"encoding/binary"
	"log"
	"net"
	"time"
)

type ApplicationGraph struct {
	graph map[udpack.IPString]*RPLLink
}

func NewApplicationGraph() ApplicationGraph {
	return ApplicationGraph{
		graph: make(map[udpack.IPString]*RPLLink),
	}
}

func (app ApplicationGraph) Type() AppType {
	return AppTypeGraph
}

func (app ApplicationGraph) ProcessPacket(_ *net.UDPAddr, packet []byte) {
    log.Println("Processing graph packet...")
	graphUpdate := decodeGraphUpdateData(packet)
	app.updateGraph(&graphUpdate)
}

func (app *ApplicationGraph) updateGraph(graphUpdate *GraphTopologyUpdate) {
	childIPString := udpack.IPString(graphUpdate.ChildIP.String())
	parentIPString := udpack.IPString(graphUpdate.ParentIP.String())
	if v, in := app.graph[childIPString]; !in || v.ParentIP != parentIPString {
		log.Printf("Adding RPL Link: from %s to %s\n", childIPString, parentIPString)
	}
	app.graph[childIPString] = &RPLLink{
		ParentIP:    parentIPString,
		ExpiredTime: time.Now().Add(time.Duration(graphUpdate.Lifetime) * time.Second),
	}
	go app.removeRPLLinkWhenLifetimeExpired(childIPString, graphUpdate.Lifetime)
}

func (app *ApplicationGraph) removeRPLLinkWhenLifetimeExpired(childIP udpack.IPString, lifetime uint32) {
	time.Sleep(time.Duration(lifetime) * time.Second)
	rplLink, in := app.graph[childIP]
	if !in || time.Now().Before(rplLink.ExpiredTime) {
		return
	}

	// Delete the child only if there is not another node that have a link to this child
	for _, link := range app.graph {
		if link.ParentIP == childIP {
			return
		}
	}
	delete(app.graph, childIP)
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

func decodeGraphUpdateData(data []byte) GraphTopologyUpdate {
	return GraphTopologyUpdate{
		ChildIP:  data[0:16],
		ParentIP: data[16:32],
		Lifetime: binary.LittleEndian.Uint32(data[32:]),
	}
}
