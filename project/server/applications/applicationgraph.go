package applications

import (
	"fmt"
	"log"
	"net"
	"scheduleupdater-server/addrtranslation"
	"scheduleupdater-server/utils"
	"sync"
	"time"
)

// ApplicationGraph retrieve the RPL graph from the nodes.
type ApplicationGraph struct {
	Graph    RPLGraph
	nClients uint
	lock     sync.RWMutex
}

func NewApplicationGraph(nClients uint) ApplicationGraph {
	return ApplicationGraph{
		Graph:    make(RPLGraph),
		nClients: nClients,
		lock:     sync.RWMutex{},
	}
}

func (app *ApplicationGraph) Type() AppType {
	return AppTypeGraph
}

func (app *ApplicationGraph) ProcessPacket(addr *net.UDPAddr, packet []byte) {
	graphUpdate := decodeGraphUpdateData(addr.IP, packet)
	app.updateGraph(&graphUpdate)
}

func (app *ApplicationGraph) Ready() bool {
	if len(app.Graph) > int(app.nClients-1) {
		utils.Log.ErrorPrintln("Something wrong...")
	}
	return len(app.Graph) == int(app.nClients-1) // -1 since we count the number of links
}

func (app *ApplicationGraph) updateGraph(graphUpdate *GraphTopologyUpdate) {
	// This is mostly a hack and should be replaced in a proper environement
	childIPString := addrtranslation.IPString(graphUpdate.ChildIP.String())
	parentIPString := addrtranslation.IPString(graphUpdate.ParentIP.String()).LinkLocalToGlobal()
	if v, in := app.Graph[childIPString]; !in || v.ParentIP != parentIPString {
		log.Printf("Adding RPL Link: from %s to %s\n", childIPString, parentIPString)
		app.Graph[childIPString] = &RPLLink{
			ParentIP: parentIPString,
		}
	}
}

type RPLGraph map[addrtranslation.IPString]*RPLLink

// LeavesToRootOrder returns an ordered list of IP addresses. The list is ordered in a way
// that the leaves are first and then the following addresses are their ancestors up to the root
// of the RPL tree.
func (rplGraph RPLGraph) LeavesToRootOrder() []addrtranslation.IPString {
	order := make([]addrtranslation.IPString, len(rplGraph)+1)
	leaves := rplGraph.findLeaves()
	index := 0

	insertIfNotIn := func(value addrtranslation.IPString) {
		for i := 0; i < index; i++ {
			if order[i] == value {
				return
			}
		}
		order[index] = value
		index++
	}

	for _, leaf := range leaves {
		insertIfNotIn(leaf)
		for {
			if link, in := rplGraph[leaf]; in {
				insertIfNotIn(link.ParentIP)
				leaf = link.ParentIP
			} else {
				break
			}
		}
	}
	fmt.Printf("ORDER: %+v", order)
	return order
}

func (rplGraph RPLGraph) findLeaves() []addrtranslation.IPString {
	leaves := make([]addrtranslation.IPString, 0)
	for potentialLeaf := range rplGraph {
		isLeaf := true
		for _, childLink := range rplGraph {
			notALeaf := childLink != nil && childLink.ParentIP == potentialLeaf
			if notALeaf {
				isLeaf = false
				break
			}
		}
		if isLeaf {
			leaves = append(leaves, potentialLeaf)
		}
	}
	return leaves
}

type RPLLink struct {
	ParentIP    addrtranslation.IPString
	ExpiredTime time.Time
}

type GraphTopologyUpdate struct {
	ChildIP  net.IP
	ParentIP net.IP
}

func decodeGraphUpdateData(addr net.IP, data []byte) GraphTopologyUpdate {
	return GraphTopologyUpdate{
		ParentIP: data[:16],
		ChildIP:  addr,
	}
}
