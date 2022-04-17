package scheduleupdater

import (
	"errors"
	"fmt"
	"log"
	"net"
	"os"
	"scheduleupdater-server/addrtranslation"
	"scheduleupdater-server/applications"
	stats "scheduleupdater-server/stats"
	"scheduleupdater-server/udpack"
	"scheduleupdater-server/utils"
	"sync"
	"time"
)

const ScheduleUpdaterPktMaxCells = 11

type Updater struct {
	conn    *udpack.UDPAckConn
	clients []addrtranslation.IPString
}

func NewUpdater(conn *udpack.UDPAckConn, clients []addrtranslation.IPString) Updater {
	utils.Log.WarningPrintln("Client len: ", len(clients))
	return Updater{
		conn:    conn,
		clients: clients,
	}
}

func (updater *Updater) UpdateClients(schedule *Schedule, rplGraph *applications.RPLGraph) {
	// helper function to check if any error and panic if any
	panicIfErrors := func(ackPackets <-chan AckPacketOrError) {
		for ackPacket := range ackPackets {
			if ackPacket.err != nil {
				log.Panic(ackPacket.err.Error())
			}
		}
	}

	// New schedule update
	stats.SimulationStats.ScheduleUpdateStart = time.Now()
	scheduleUpdateAckPackets := updater.sendToEachClientAsync(schedule.Serialize, updater.clients)
	for ackPacket := range scheduleUpdateAckPackets {
		if ackPacket.err != nil {
			log.Panic(ackPacket.err.Error())
		}
		if ackPacket.confirmation() == AckPacketConfirmationDecline {
			// TODO HANDLE decline
			panic("One node declined the new schedule, the code does not handle yet this case")
		}
	}
	log.Println("No errors detected while sending the new schedule 🎉")

	order := rplGraph.LeavesToRootOrder()
	// Update complete
	updateCompleteErrors := updater.sendToEachClientSync(func(clientIP addrtranslation.IPString) ([][]byte, error) {
		updateCompletePkt := UpdateConfirmation{}
		return [][]byte{updateCompletePkt.Encode()}, nil
	}, order)
	stats.SimulationStats.ScheduleUpdateEnd = time.Now()
	panicIfErrors(updateCompleteErrors)
	log.Println("No errors detected while sending complete pkt 🎉")
	log.Println("Everything is ok don't worry! Be happy 🎉🎉🎉")
	stats.SimulationStats.WriteToFile("stats")
	os.Exit(0)
}

type Serializer = func(clientIP addrtranslation.IPString) ([][]byte, error)

func (updater *Updater) sendToEachClientAsync(serialize Serializer, order []addrtranslation.IPString) <-chan AckPacketOrError {
	var wg sync.WaitGroup
	//clientCount := len(updater.clients)
	ackPackets := make(chan AckPacketOrError, 10000) // TODO MODIFY THIS SOULD NOT TAKE 1000 entries
	for _, clientIP := range updater.clients {
		wg.Add(1)
		go updater.serializeAndSend(clientIP, serialize, ackPackets, &wg)
	}
	wg.Wait()
	close(ackPackets)
	return ackPackets
}

func (updater *Updater) sendToEachClientSync(serialize Serializer, order []addrtranslation.IPString) <-chan AckPacketOrError {
	var wg sync.WaitGroup
	//clientCount := len(updater.clients)
	ackPackets := make(chan AckPacketOrError, 10000) // TODO MODIFY THIS SOULD NOT TAKE 1000 entries
	for _, clientIP := range updater.clients {
		wg.Add(1)
		updater.serializeAndSend(clientIP, serialize, ackPackets, &wg)
	}
	wg.Wait()
	close(ackPackets)
	return ackPackets
}

type AckPacketOrError struct {
	err    error
	packet []byte
}

type AckPacketConfirmation byte

const (
	AckPacketConfirmationDecline = iota
	AckPacketConfirmationOK
)

func (ackPacket *AckPacketOrError) confirmation() AckPacketConfirmation {
	if ackPacket.packet == nil {
		utils.Log.ErrorPrintln("AckPacket does not contain a packet")
		return AckPacketConfirmationDecline
	}
	return AckPacketConfirmation(ackPacket.packet[0])
}

func (updater *Updater) serializeAndSend(clientIP addrtranslation.IPString, serialize Serializer, ackPackets chan AckPacketOrError, wg *sync.WaitGroup) {
	defer wg.Done()

	pkts, err := serialize(clientIP)
	if err != nil {
		ackPackets <- AckPacketOrError{err: err}
		return
	}
	for i, pkt := range pkts {
		utils.Log.Println("Sending to client: ", clientIP, ", packet ", i, " / ", len(pkts))
		// @incomplete needs refactor
		udpAddr := &net.UDPAddr{
			IP:   net.ParseIP(string(clientIP)),
			Port: 8765,
			Zone: "",
		}
		err, packet := updater.conn.WriteTo(pkt, udpAddr)
		if err != nil {
			utils.Log.ErrorPrintln(err.Error())
			ackPackets <- AckPacketOrError{err: err}
			return
		}
		ackPackets <- AckPacketOrError{packet: packet}
	}
}

type LinkOptions uint8

const (
	LinkOptionTX          = 1
	LinkOptionRX          = 2
	LinkOptionShared      = 4
	LinkOptionTimeKeeping = 8
)

type Cell struct {
	LinkOptions LinkOptions
	TimeSlot    uint16
	Channel     uint16
}

func (cell *Cell) Equals(other *Cell) bool {
	return cell.TimeSlot == other.TimeSlot &&
		cell.Channel == other.Channel
}

type PktType int

const (
	PktTypeUpdateRequest = iota
	PktTypeUpdateConfirmation
)

type UpdateRequest struct {
	NeighborAddr addrtranslation.MacAddr
	Cells        []Cell
}

func (pkt *UpdateRequest) Type() PktType {
	return PktTypeUpdateRequest
}

func (pkt *UpdateRequest) Encode() []byte {
	buffer := make([]byte, 0)
	buffer = append(buffer, uint8(pkt.Type()))
	// Appending the neighborAddr uint16 to the buffer
	for _, v := range pkt.NeighborAddr {
		buffer = append(buffer, v)
	}
	buffer = append(buffer, uint8(len(pkt.Cells)))
	for i := 0; i < len(pkt.Cells); i++ {
		buffer = append(buffer, uint8(pkt.Cells[i].LinkOptions))
		buffer = utils.AppendLittleEndianUint16(buffer, pkt.Cells[i].TimeSlot)
		buffer = utils.AppendLittleEndianUint16(buffer, pkt.Cells[i].Channel)
	}
	return buffer
}

type UpdateConfirmation struct{}

func (pkt *UpdateConfirmation) Type() PktType {
	return PktTypeUpdateConfirmation
}

func (pkt *UpdateConfirmation) Encode() []byte {
	return []byte{uint8(pkt.Type())}
}

type Schedule map[addrtranslation.IPString]map[*addrtranslation.MacAddr][]Cell

func NewSchedule() Schedule {
	return make(Schedule)
}

func (schedulePtr *Schedule) AddCell(nodeAddr addrtranslation.IPString, neighborAddr *addrtranslation.MacAddr, cell *Cell) {
	schedule := *schedulePtr
	mapCells, in := schedule[nodeAddr]
	if !in {
		mapCells = make(map[*addrtranslation.MacAddr][]Cell)
		schedule[nodeAddr] = mapCells
	}
	cells, in := mapCells[neighborAddr]
	if !in {
		cells = []Cell{}
		mapCells[neighborAddr] = cells
	}
	mapCells[neighborAddr] = append(cells, *cell)
}

func (schedulePtr Schedule) Serialize(clientIP addrtranslation.IPString) ([][]byte, error) {
	clientSchedule, in := schedulePtr[clientIP]
	if !in {
		return nil, errors.New(fmt.Sprintf("the udpack ip address (%s) doesn't have any schedule associated with it", clientIP))
	}

	pkts := make([][]byte, 0)
	for neighborAddr, neighborCells := range clientSchedule {
		for i := 0; i < len(neighborCells); i += ScheduleUpdaterPktMaxCells {
			pkt := UpdateRequest{
				NeighborAddr: *neighborAddr,
				Cells:        neighborCells[i:min(i+ScheduleUpdaterPktMaxCells, len(neighborCells))],
			}
			pkts = append(pkts, pkt.Encode())
		}
	}

	log.Println("Number of pkts to send: ", len(pkts))
	return pkts, nil
}

func (schedulePtr Schedule) IsCellUsed(nodeAddr addrtranslation.IPString, neighborAddr *addrtranslation.MacAddr, cell *Cell) bool {
	for _, scheduleCell := range schedulePtr[nodeAddr][neighborAddr] {
		if cell.Equals(&scheduleCell) {
			return true
		}
	}
	return false
}

// ----- INTERNAL ----

func min(x, y int) int {
	if x <= y {
		return x
	}
	return y
}
