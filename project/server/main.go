package main

import (
	"coap-server/applications"
	"coap-server/scheduleupdater"
	"coap-server/udpack"
	"errors"
	"fmt"
	"log"
	"net"
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
            if (appGraphReady && appBandwidthReady) {
                break
            }
            if (!appGraphReady) {
                fmt.Println("App Graph is not ready yet...")
            }
            if (!appBandwidthReady) {
                fmt.Println("App Bandwidth is not ready yet...")
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

func addCell(s *scheduleupdater.Schedule, addr udpack.IPString, neighborAddr udpack.IPString, timeslot uint16, channel uint16) {
	s.AddCell(addr,
		neighborAddr,
		&scheduleupdater.Cell{
			LinkOptions: 1,
			TimeSlot:    timeslot,
			Channel:     channel,
		})
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
		LinkOptions: 1,
		TimeSlot:    0,
		Channel:     0,
	}
	txCell := scheduleupdater.Cell{
		LinkOptions: 0,
		TimeSlot:    0,
		Channel:     0,
	}
	for timeslot := uint16(0); timeslot < 14; timeslot++ {
		for channel := uint16(0); channel < 16; channel++ {
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
	return errors.New("No available cell left")
}
