package main

import (
	"coap-server/applications"
	"coap-server/scheduleupdater"
	"coap-server/udpack"
	"coap-server/utils"
	"errors"
	"fmt"
	"log"
	"net"
	"os"
	"strconv"
	"time"
)

func printHelp() {
	fmt.Println("server [#MOTES] [PORT]")
	fmt.Println("   - #MOTES the number of motes in the simulation including the border router")
	fmt.Println("   - PORT the server port")
}

func main() {
	utils.NewLogger(utils.LogLevelInfo, utils.WHITE)
	nClients, port, err := parseArgs()
	if err != nil {
		fmt.Println(err)
		printHelp()
		os.Exit(1)
	}
	addr := &net.UDPAddr{
		Port: port,
		IP:   net.ParseIP("0.0.0.0"),
	}
	conn, err := net.ListenUDP("udp", addr)
	if err != nil {
		log.Panic(err)
	}

	server := udpack.NewUDPAckServer(conn, nil)
	appGraph := applications.NewApplicationGraph(nClients)
	appBandwidth := applications.NewApplicationBandwidth(nClients)
	appTopology := applications.NewApplicationTopology(nClients)
	appDispatcher := applications.NewAppDispatcher().
		Subscribe(&appGraph).
		Subscribe(&appBandwidth).
		Subscribe(applications.NewApplicationHelloWorld()).
		Subscribe(&appTopology)

	addrs := initializeClientsAddrs(nClients)
	go func() {
		for {
			appGraphReady := appGraph.Ready()
			appBandwidthReady := appBandwidth.Ready()
			if appGraphReady && appBandwidthReady {
				break
			}
			if !appGraphReady {
				utils.Log.WarningPrintln("App Graph is not ready yet...")
			}
			if !appBandwidthReady {
				utils.Log.WarningPrintln("App Bandwidth is not ready yet...")
			}
			time.Sleep(time.Second * 4)
		}
		schedule := generateSchedule(&appGraph.Graph, &appBandwidth.Bandwith)
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

func initializeClientsAddrs(clientCount uint) []udpack.IPString {
	addrs := make([]udpack.IPString, clientCount)
	for i := uint(1); i < clientCount+1; i++ {
		fmt.Printf("Adding fd00::20%d:%d:%d:%d\n", i, i, i, i)
		addrs[i-1] = udpack.IPString(fmt.Sprintf("fd00::20%d:%d:%d:%d", i, i, i, i))
	}
	return addrs
}

func generateSchedule(graph *applications.RPLGraph, bandwidthMap *applications.BandwidthMap) scheduleupdater.Schedule {
	schedule := scheduleupdater.NewSchedule()
	for mote, bandwidth := range *bandwidthMap {
		for i := uint(0); i < bandwidth; i++ {
			if rplLink, in := (*graph)[mote]; in {
				err := addOneCell(&schedule, mote, rplLink.ParentIP)
				if err != nil {
					log.Panic(err)
				}
			}
		}
	}
	return schedule
}

func addOneCell(schedule *scheduleupdater.Schedule, mote udpack.IPString, neighbor udpack.IPString) error {
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
		for timeslot := uint16(1); timeslot < 14; timeslot++ {
			rxCell.TimeSlot = timeslot
			rxCell.Channel = channel
			txCell.TimeSlot = timeslot
			txCell.Channel = channel

			if !schedule.IsCellUsed(mote, neighbor, &txCell) && !schedule.IsCellUsed(neighbor, mote, &rxCell) {
				schedule.AddCell(mote, neighbor, &txCell)
				schedule.AddCell(neighbor, mote, &rxCell)
				return nil
			}
		}
	}
	return errors.New("no available cell left")
}

func parseArgs() (uint, int, error) {
	if len(os.Args) != 3 {
		return 0, 0, errors.New("wrong command line usage")
	}
	nClients, err := strconv.Atoi(os.Args[1])
	if err != nil {
		return 0, 0, err
	}
	port, err := strconv.Atoi(os.Args[2])
	if err != nil {
		return 0, 0, err
	}
	return uint(nClients), port, nil
}
