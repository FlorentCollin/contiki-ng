package main

import (
	"coap-server/scheduleupdater"
	"encoding/binary"
	"fmt"
	"log"
	"net"
	"time"
)

const nClients = 5

func initializeSchedule(clientCount uint) scheduleupdater.Schedule {
	schedule := scheduleupdater.NewSchedule()
	addrs := make([]string, clientCount)
	for i := uint(2); i < clientCount+2; i++ {
		fmt.Printf("Adding fd00::20%d:%d:%d:%d\n", i, i, i, i)
		addrs[i-2] = fmt.Sprintf("fd00::20%d:%d:%d:%d", i, i, i, i)
	}
	log.Println("Number of clients: ", clientCount)
	for i := uint16(1); i < 2; i++ {
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

func initializeClientsStates(conn *net.UDPConn, clientsChannels map[string]chan []byte) map[string]*ClientState {
    clientsStates := make(map[string]*ClientState)
    for addr, c := range clientsChannels {
        udpAddr := net.UDPAddr{
        	IP:   net.ParseIP(addr),
        	Port: 8765,
        	Zone: "",
        }
        clientState := NewClientState(conn, &udpAddr, c)
        clientsStates[addr] = &clientState
    }
    return clientsStates
}

func SendResponse(conn *net.UDPConn, addr *net.UDPAddr, msg []byte) {
	_, err := conn.WriteToUDP(msg, addr)
	if err != nil {
		log.Panicln(err)
	}
}

type ClientState struct {
    conn       *net.UDPConn
    addr       *net.UDPAddr
	pktNumber  uint16
	pktChan    <-chan []byte
}

func NewClientState(conn *net.UDPConn, addr *net.UDPAddr, pktChan <-chan []byte) ClientState {
    return ClientState {
        conn: conn,
        addr: addr,
        pktNumber: 1,
        pktChan: pktChan,
    }
}

func (client *ClientState) send(pktToSend []byte) {
    for {
        client.conn.WriteTo(pktToSend, client.addr)
        select {
        case pkt := <- client.pktChan:
            var pktNumber uint16 = (uint16(pkt[0]) << 8) + uint16(pkt[1])
            log.Printf("Expecting ACK: %d, got %d\n", client.pktNumber, pktNumber)
            if pktNumber == client.pktNumber {
                log.Println("ACK correctly received")
                fmt.Println()
                client.pktNumber++
                return; // Client correctly received the pktToSend
            } else if pktNumber < client.pktNumber {
                log.Println("Already received this ack -> doing nothing");
            } else {
                log.Println("Message received that was not the expected message, resending pktToSend")
                log.Println("Message received: ", string(pkt))
            }
        case <-time.After(15 * time.Second):
            log.Printf("Timeout on addr %s, resending pkt\n", client.addr.IP)
        }
    }
}

func (client *ClientState) periodicSend() {
    for {
        buffer := make([]byte, 2)
        binary.BigEndian.PutUint16(buffer, client.pktNumber)
        client.send(buffer)
        time.Sleep(time.Millisecond * 500)
    } 
}

func main() {
	schedule := initializeSchedule(nClients)
	clientsChannels := initializeChannels(&schedule)

	conn, err := net.ListenUDP("udp", &net.UDPAddr{
		Port: 3000,
		IP:   net.ParseIP("0.0.0.0"),
	})
	if err != nil {
		panic(err)
	}
    clientsStates := initializeClientsStates(conn, clientsChannels)
    for _, clientState := range clientsStates {
        go clientState.periodicSend()
    }


	defer conn.Close()
	fmt.Printf("server listening %s\n", conn.LocalAddr().String())

	message := make([]byte, 2048)
	for {
		rlen, remote, err := conn.ReadFromUDP(message[:])
		if err != nil {
			panic(err)
		}
		c, in := clientsChannels[remote.IP.String()]
		if !in {
			log.Panic("Received a message from an address that is not in the schedule: ", remote.IP.String())
		}
        log.Println("Message received from ", remote.IP.String())
		data := message[:rlen]
		c <- data
	}
}
