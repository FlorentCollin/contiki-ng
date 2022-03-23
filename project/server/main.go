package main

import (
	"coap-server/applications"
	"coap-server/scheduleupdater"
	"coap-server/udpack"
	"fmt"
	"log"
	"net"
	"os"
	"os/exec"
	"strconv"
	"strings"
	"time"
)

//const nCells = 10
const nClients = 2

func main() {
	addr := &net.UDPAddr{
		Port: 3000,
		IP:   net.ParseIP("0.0.0.0"),
	}
	conn, err := net.ListenUDP("udp", addr)
	if err != nil {
		log.Panic(err)
	}

	server := udpack.NewUDPAckServer(conn, nil)
	appGraph := applications.NewApplicationGraph(nClients)
	appBandwith := applications.NewApplicationBandwith(nClients)
	appTopology := applications.NewApplicationTopology(nClients)
	appDispatcher := applications.NewAppDispatcher().
		Subscribe(&appGraph).
		Subscribe(&appBandwith).
		Subscribe(applications.NewApplicationHelloWorld()).
		Subscribe(&appTopology)

	addrs := initializeClientsAddrs(nClients)
	go func() {
		schedule := initializeSchedule(addrs, &appGraph, &appTopology, &appBandwith)
		updater := scheduleupdater.NewUpdater(server, addrs)
		updater.UpdateClients(&schedule)
	}()

	defer func(server *udpack.UDPAckConn) {
		err := server.Close()
		if err != nil {
			log.Panic(err)
		}
	}(server)
	fmt.Println("server listening ...")
	log.Panic(server.Serve(appDispatcher.Handler))
}

// -- INTERNAL --

func shellSchedule(schedule *scheduleupdater.Schedule, addrs []udpack.IPString) {
	cmd := exec.Command("clingo", "-t 8", "facts.pl", "main.pl")
	stdout, _ := cmd.Output()
	parseOutput(schedule, addrs, string(stdout))
}

func parseOutput(schedule *scheduleupdater.Schedule, addrs []udpack.IPString, output string) {
	for _, line := range strings.Split(output, "\n") {
		if !strings.HasPrefix(line, "cell") {
			continue
		}
		for _, cell := range strings.Split(line, " ") {
			parameters := strings.Split(cell[5:len(cell)-1], ",")
			n, err := strconv.ParseInt(parameters[0], 10, 32)
			if err != nil {
				log.Panic(err.Error())
			}
			m, err := strconv.ParseInt(parameters[1], 10, 32)
			if err != nil {
				log.Panic(err.Error())
			}
			s, err := strconv.ParseUint(parameters[2], 10, 16)
			if err != nil {
				log.Panic(err.Error())
			}
			c, err := strconv.ParseUint(parameters[3], 10, 16)
			if err != nil {
				log.Panic(err.Error())
			}
			fmt.Println(n, s, c)
			addCell(schedule, addrs[n], addrs[m], uint16(s-1), uint16(c-1))
		}
	}
}

func initializeClientsAddrs(clientCount uint) []udpack.IPString {
	addrs := make([]udpack.IPString, clientCount)
	for i := uint(1); i < clientCount+1; i++ {
		fmt.Printf("Adding fd00::20%d:%d:%d:%d\n", i, i, i, i)
		addrs[i-1] = udpack.IPString(fmt.Sprintf("fd00::20%d:%d:%d:%d", i, i, i, i))
	}
	return addrs
}

func initializeSchedule(
	addrs []udpack.IPString,
	appGraph *applications.ApplicationGraph,
	appTopology *applications.ApplicationTopology,
	appBandwidth *applications.ApplicationBandwith,
) scheduleupdater.Schedule {

	for !appGraph.Ready() || !appTopology.Ready() || !appBandwidth.Ready() {
		if appGraph.Ready() {
			log.Println("GRAPH IS READY HOUPI")
		}
		if appTopology.Ready() {
			log.Println("TOPOLOGY IS READY HOUPI")
		}
		if appBandwidth.Ready() {
			log.Println("BANDWIDTH IS READY HOUPI")
		}
		time.Sleep(3 * time.Second)
	}
	log.Println("Everything is ready create the new schedule")
	generateFacts(addrs, &appGraph.Graph, &appTopology.Topology, &appBandwidth.Bandwith)
	schedule := scheduleupdater.NewSchedule()
	shellSchedule(&schedule, addrs)
	return schedule
}

func generateFacts(addrs []udpack.IPString, graph *applications.RPLGraph, topology *applications.Topology, bandwidthMap *applications.BandwidthMap) {
	f, err := os.Create("facts.pl")
	if err != nil {
		log.Fatal(err)
	}
	defer func(f *os.File) {
		err := f.Close()
		if err != nil {
			log.Panic(err.Error())
		}
	}(f)

	writeStringPanic(f, fmt.Sprintf("#const nNodes = %d.\n", nClients-1))
	writeStringPanic(f, "#const nSlots = 21.\n")
	writeStringPanic(f, "#const nChannels = 16.\n")
	writeStringPanic(f, "\n")
	writeStringPanic(f, "node(0..nNodes).\n")

	for n, link := range *graph {
		m := link.ParentIP
		indexN := indexOf(addrs, n)
		indexM := indexOf(addrs, m)
		writeStringPanic(f, fmt.Sprintf("graph(%d, %d). ", indexN, indexM))
	}
	writeStringPanic(f, "\n")

    // TODO: This part is mostly non sense since we can't deduce the topology just by enumerating
    // all neighbors of a node. So we remove it from now on.
	//for n, neighbors := range *topology {
		//for _, m := range neighbors {
			//indexN := indexOf(addrs, n)
			//indexM := indexOf(addrs, m)
			//writeStringPanic(f, fmt.Sprintf("topology(%d, %d). ", indexN, indexM))
		//}
	//}
	writeStringPanic(f, "\n")

	totalBandwidth := uint(0)
	for n, bandwidth := range *bandwidthMap {
		totalBandwidth += bandwidth
		indexN := indexOf(addrs, n)
		writeStringPanic(f, fmt.Sprintf("bandwidth(%d, %d). ", indexN, bandwidth))
	}
	writeStringPanic(f, "\n")
	writeStringPanic(f, fmt.Sprintf("totalBandwidth(%d).\n", totalBandwidth))
}

func writeStringPanic(f *os.File, str string) {
	_, err := f.WriteString(str)
	if err != nil {
		log.Panic(err.Error())
	}
}

func indexOf(addrs []udpack.IPString, addr udpack.IPString) int {
	for i, v := range addrs {
		if addr == v {
			return i
		}
	}
	log.Panic(fmt.Sprintf("Addr not found in array: %s", addr))
	return -1
}

func addCell(s *scheduleupdater.Schedule, addr udpack.IPString, neighbor udpack.IPString, timeslot uint16, channel uint16) {
	ip := net.ParseIP(string(neighbor))
	var neighborAddr scheduleupdater.NeighborAddr
	copy(neighborAddr[:], ip[:])
	s.AddCell(addr,
		neighborAddr,
		&scheduleupdater.Cell{
			LinkOptions: 1,
			TimeSlot:    timeslot,
			Channel:     channel,
		})
}
