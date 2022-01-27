package main

import (
	"coap-server/client"
	"coap-server/scheduleupdater"
	"fmt"
	"log"
	"net"
)

const nClients = 8
const nCells = 20

func main() {
	schedule := initializeSchedule(nClients)
	clientsChannels := initializeChannels(&schedule)

	conn, err := net.ListenUDP("udp", &net.UDPAddr{
		Port: 3000,
		IP:   net.ParseIP("0.0.0.0"),
	})
	if err != nil {
		log.Panic(err)
	}
	clientsStates := initializeClientsStates(conn, clientsChannels)
	updater := scheduleupdater.NewUpdater(clientsStates)
	go updater.UpdateClients(&schedule)

	defer func(conn *net.UDPConn) {
		err := conn.Close()
		if err != nil {
			log.Panic(err)
		}
	}(conn)
	fmt.Printf("server listening %s\n", conn.LocalAddr().String())

	message := make([]byte, 2048)
	for {
		rlen, remote, err := conn.ReadFromUDP(message[:])
		log.Println("Trigger")
		if err != nil {
			log.Panic(err)
		}
		c, in := clientsChannels[remote.IP.String()]
		if !in {
			log.Panic("Received a message from an address that is not in the schedule: ", remote.IP.String())
		}
		log.Println("Message received from ", remote.IP.String())
		data := message[:rlen]
		go func() {
			select {
			case c <- data:
			default:
				log.Println("Discarding packet no goroutine requested it")
			}
		}()
	}
}

// -- INTERNAL --

func initializeSchedule(clientCount uint) scheduleupdater.Schedule {
	schedule := scheduleupdater.NewSchedule()
	addrs := make([]string, clientCount)
	for i := uint(2); i < clientCount+2; i++ {
		fmt.Printf("Adding fd00::20%d:%d:%d:%d\n", i, i, i, i)
		addrs[i-2] = fmt.Sprintf("fd00::20%d:%d:%d:%d", i, i, i, i)
	}
	log.Println("Number of clients: ", clientCount)
	for i := uint16(1); i < nCells+1; i++ {
		for _, addr := range addrs {
			addCell(&schedule, addr, i, i+1)
		}
	}
	log.Println("Schedule len : ", len(schedule))
	return schedule
}

func addCell(s *scheduleupdater.Schedule, addr string, timeslot uint16, channel uint16) {
	neighborAddr := [4]uint16{0xf80, 0x202, 0x2, 0x2}
	s.AddCell(addr,
		neighborAddr,
		&scheduleupdater.Cell{
			LinkOptions: 1,
			TimeSlot:    timeslot,
			Channel:     channel,
		})
}

func initializeChannels(s *scheduleupdater.Schedule) map[string]chan []byte {
	channels := make(map[string]chan []byte)
	for addr := range *s {
		channels[addr] = make(chan []byte)
	}
	return channels
}

func initializeClientsStates(conn *net.UDPConn, clientsChannels map[string]chan []byte) []*client.State {
	clientsStates := make([]*client.State, 0)
	for addr, c := range clientsChannels {
		udpAddr := net.UDPAddr{
			IP:   net.ParseIP(addr),
			Port: 8765,
			Zone: "",
		}
		clientState := client.NewClientState(conn, &udpAddr, c)
		clientsStates = append(clientsStates, &clientState)
	}
	return clientsStates
}
