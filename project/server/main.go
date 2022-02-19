package main

import (
	"coap-server/applications"
	"coap-server/scheduleupdater"
	"coap-server/udpack"
	"fmt"
	"log"
	"net"
	"os/exec"
	"strings"
    "strconv"
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
	appGraph := applications.NewApplicationGraph()
	appBandwith := applications.NewApplicationBandwith()
	appTopology := applications.NewApplicationTopology()
	appDispatcher := applications.NewAppDispatcher().
		Subscribe(appGraph).
		Subscribe(appBandwith).
		Subscribe(applications.NewApplicationHelloWorld()).
		Subscribe(appTopology)

	schedule, clients := initializeSchedule(nClients)
	updater := scheduleupdater.NewUpdater(server, clients)
	go updater.UpdateClients(&schedule)

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
    generator := exec.Command("python", "generator.py", fmt.Sprint(nClients), "5", "5", "10")
    generator.Run()
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
            //m, err := strconv.ParseInt(parameters[1], 10, 32)
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
            addCell(schedule, addrs[n], uint16(s-1), uint16(c-1))
        }
    }
}

func initializeSchedule(clientCount uint) (scheduleupdater.Schedule, []udpack.IPString) {
	schedule := scheduleupdater.NewSchedule()
	addrs := make([]udpack.IPString, clientCount+1)
	for i := uint(1); i < clientCount+2; i++ {
		fmt.Printf("Adding fd00::20%d:%d:%d:%d\n", i, i, i, i)
		addrs[i-1] = udpack.IPString(fmt.Sprintf("fd00::20%d:%d:%d:%d", i, i, i, i))
	}
	log.Println("Number of clients: ", clientCount)
    shellSchedule(&schedule, addrs)
	//for i := uint16(1); i < nCells+1; i++ {
		//for _, addr := range addrs {
			//addCell(&schedule, addr, i, i+1)
		//}
	//}
	return schedule, addrs
}

func addCell(s *scheduleupdater.Schedule, addr udpack.IPString, timeslot uint16, channel uint16) {
	neighborAddr := [4]uint16{0xf80, 0x202, 0x2, 0x2}
	s.AddCell(addr,
		neighborAddr,
		&scheduleupdater.Cell{
			LinkOptions: 1,
			TimeSlot:    timeslot,
			Channel:     channel,
		})
}

//func initializeChannels(s *scheduleupdater.Schedule) map[string]chan []byte {
//	channels := make(map[string]chan []byte)
//	for addr := range *s {
//		channels[addr] = make(chan []byte)
//	}
//	return channels
//}
//
//func initializeClientsStates(conn *net.UDPConn, clientsChannels map[string]chan []byte) []*udpack.UDPAckClient {
//	clientsStates := make([]*udpack.UDPAckClient, 0)
//	for addr, c := range clientsChannels {
//		udpAddr := net.UDPAddr{
//			IP:   net.ParseIP(addr),
//			Port: 8765,
//			Zone: "",
//		}
//		clientState := udpack.NewClientStateWithACK(conn, &udpAddr, c)
//		clientsStates = append(clientsStates, &clientState)
//	}
//	return clientsStates
//}
