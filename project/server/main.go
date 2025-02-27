package main

// Server main application.
// This application waits for a number of nodes to connect to the server.
// The nodes are responsible to send three information:
// 1. Their bandwidth need
// 2. Their neighbors
// 3. Thei RPL parent
//
// When all these three information have been sent to the server,
// the server generate a new schedule based on a greedy algorithm.
// This newly generated schedule is then sent to the node based on the protocole
// described in the master's thesis PDF.
// As the goal was only to install a new schedule once for the simulation,
// the server stops when the schedule was correctly sent to the nodes.

import (
	"errors"
	"fmt"
	"log"
	"net"
	"os"
	"scheduleupdater-server/addrtranslation"
	"scheduleupdater-server/applications"
	"scheduleupdater-server/scheduleupdater"
	"scheduleupdater-server/stats"
	"scheduleupdater-server/udpack"
	"scheduleupdater-server/utils"
	"strconv"
	"time"
)

func printHelp() {
	fmt.Println("server [#MOTES] [FIRST_MOTE_ID] [PORT]")
	fmt.Println("   - #MOTES the number of motes in the simulation including the border router")
	fmt.Println("   - FIRST_MOTE_ID the id of the first mote which is the border router")
	fmt.Println("   - PORT the server port")
}

func main() {
	utils.NewLogger(utils.LogLevelInfo, utils.WHITE)

	nClients, firstMoteID, port, timeout, err := parseArgs()
	if err != nil {
		fmt.Println(err)
		printHelp()
		os.Exit(1)
	}
	stats.SimulationStats.Nclients = nClients

	addr := &net.UDPAddr{
		Port: port,
		IP:   net.ParseIP("0.0.0.0"),
	}

	// Start the udp listener
	conn, err := net.ListenUDP("udp", addr)
	if err != nil {
		log.Panic(err)
	}

	config := udpack.UDPAckConnSendConfig{
		MaxRetries:          100,
		TimesBetweenRetries: 1,
		Timeout:             time.Duration(timeout) * time.Second,
	}
	server := udpack.NewUDPAckServer(conn, &config)
	stats.SimulationStats.Timeout = server.Config.Timeout.Seconds()

	// Create the three application, when a node send a packet,
	// it is correctly dispatch to the application based on the ApplicationType
	appGraph := applications.NewApplicationGraph(nClients)
	appBandwidth := applications.NewApplicationBandwidth(nClients)
	appTopology := applications.NewApplicationTopology(nClients)
	appDispatcher := applications.NewAppDispatcher().
		Subscribe(&appGraph).
		Subscribe(&appBandwidth).
		Subscribe(applications.NewApplicationHelloWorld()).
		Subscribe(&appTopology)

	// We assume that we know the address of all the nodes in the network.
	// However, we could also ask the server to keep in memory the last addresses
	// from which we received a packet.
	addrs := initializeClientsAddrs(nClients, firstMoteID)
	go func() {
		for {
			// Wait till all applications are ready
			appGraphReady := appGraph.Ready()
			appBandwidthReady := appBandwidth.Ready()
			appTopologyReady := appTopology.Ready()
			if appGraphReady && appBandwidthReady && appTopologyReady {
				break
			}
			if !appGraphReady {
				utils.Log.WarningPrintln("App Graph is not ready yet...")
			}
			if !appBandwidthReady {
				utils.Log.WarningPrintln("App Bandwidth is not ready yet...")
			}
			if !appTopologyReady {
				utils.Log.WarningPrintln("App Topology is not ready yet...")
			}
			time.Sleep(time.Second * 4)
		}
		// Generate a new schedule and send it to the nodes
		schedule := generateSchedule(&appGraph.Graph, &appBandwidth.Bandwith, &appTopology.Topology)
		updater := scheduleupdater.NewUpdater(server, addrs)
		updater.UpdateClients(&schedule, &appGraph.Graph)
	}()

	defer func(server *udpack.UDPAckConn) {
		err := server.Close()
		if err != nil {
			log.Panic(err)
		}
	}(server)
	fmt.Println("server listening on port ", port, "...")
	log.Panic(server.Serve(appDispatcher.Handler))
}

// -- INTERNAL --
// The functions below are helper to generate a new schedule based on a greedy algorithm.
// In a real use case, a good centralized scheduler like TASA should be used.

func initializeClientsAddrs(clientCount uint, firstMoteID uint) []addrtranslation.IPString {
	addrs := make([]addrtranslation.IPString, clientCount)
	for i := firstMoteID; i < firstMoteID+clientCount; i++ {
		fmt.Printf("Adding fd00::%x:%x:%x:%x\n", 512+i, i, i, i)
		addrs[i-firstMoteID] = addrtranslation.IPString(fmt.Sprintf("fd00::%x:%x:%x:%x", 512+i, i, i, i))
	}
	return addrs
}

func generateSchedule(graph *applications.RPLGraph, bandwidthMap *applications.BandwidthMap, topology *applications.Topology) scheduleupdater.Schedule {
	schedule := scheduleupdater.NewSchedule()
	for mote, bandwidth := range *bandwidthMap {
		rplLink, in := (*graph)[mote]
		if !in {
			utils.Log.ErrorPrintln("Couldn't found the RPL Link associated with mote ", mote)
			continue
		}
		// Add the descending cell to join the node from the server
		err := addOneCell(&schedule, rplLink.ParentIP, mote, topology)
		if err != nil {
			log.Panic(err)
		}
		for i := uint(0); i < bandwidth; i++ {
			err := addOneCell(&schedule, mote, rplLink.ParentIP, topology)
			if err != nil {
				log.Panic(err)
			}
		}
	}
	return schedule
}

func addOneCell(schedule *scheduleupdater.Schedule, mote addrtranslation.IPString, neighbor addrtranslation.IPString, topology *applications.Topology) error {
	rxCell := scheduleupdater.Cell{
		LinkOptions: scheduleupdater.LinkOptionRX,
		TimeSlot:    0,
		Channel:     0,
	}
	txCell := scheduleupdater.Cell{
		LinkOptions: scheduleupdater.LinkOptionTX,
		TimeSlot:    0,
		Channel:     0,
	}
	for channel := uint16(1); channel < 16; channel++ {
		for timeslot := uint16(1); timeslot < 101; timeslot++ {
			rxCell.TimeSlot = timeslot
			rxCell.Channel = channel
			txCell.TimeSlot = timeslot
			txCell.Channel = channel

			cellIsFree := true
			for _, macNeighbor := range topology.TopologyMap[mote] {
				if schedule.IsCellUsed(mote, macNeighbor, &txCell) {
					cellIsFree = false
				}
			}
			macNeighbor, ok := topology.MacIPTranslation.FindMac(neighbor)
			if !ok {
				return errors.New("could not find the mac address associated with the neighbor")
			}
			macMote, ok := topology.MacIPTranslation.FindMac(mote)
			if !ok {
				return errors.New("could not find the mac address associated with the current mote")
			}
			if cellIsFree {
				schedule.AddCell(mote, macNeighbor, &txCell)
				schedule.AddCell(neighbor, macMote, &rxCell)
				return nil
			}
		}
	}
	return errors.New("no available cell left")
}

func parseArgs() (uint, uint, int, int, error) {
	if len(os.Args) != 5 {
		return 0, 0, 0, 0, errors.New("wrong command line usage")
	}
	nClients, err := strconv.Atoi(os.Args[1])
	if err != nil {
		return 0, 0, 0, 0, err
	}
	firstMoteID, err := strconv.Atoi(os.Args[2])
	if err != nil {
		return 0, 0, 0, 0, err
	}
	port, err := strconv.Atoi(os.Args[3])
	if err != nil {
		return 0, 0, 0, 0, err
	}
	timeout, err := strconv.Atoi(os.Args[4])
	if err != nil {
		return 0, 0, 0, 0, err
	}
	return uint(nClients), uint(firstMoteID), port, timeout, nil
}
