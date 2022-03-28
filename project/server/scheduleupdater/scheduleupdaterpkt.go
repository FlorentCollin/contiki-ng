package scheduleupdater

import (
	"coap-server/udpack"
	"coap-server/utils"
	"errors"
	"fmt"
	"log"
	"net"
	"os"
	"sync"
)

const ScheduleUpdaterPktMaxCells = 11

type Updater struct {
	conn    *udpack.UDPAckConn
	clients []udpack.IPString
}

func NewUpdater(conn *udpack.UDPAckConn, clients []udpack.IPString) Updater {
	return Updater{
		conn:    conn,
		clients: clients,
	}
}

func (updater *Updater) UpdateClients(schedule *Schedule) {
	// helper function to check if any error and panic if any
	panicIfErrors := func(errs <-chan error) {
		for err := range errs {
			if err != nil {
				log.Panic(err.Error())
			}
		}
	}

	// New schedule update
	scheduleUpdateErrors := updater.sendToEachClient(schedule.Serialize)
	panicIfErrors(scheduleUpdateErrors)
	log.Println("No errors detected while sending the new schedule 🎉")

	// Update complete
	updateCompleteErrors := updater.sendToEachClient(func(clientIP udpack.IPString) ([][]byte, error) {
		updateCompletePkt := UpdateConfirmation{}
		return [][]byte{updateCompletePkt.Encode()}, nil
	})
	panicIfErrors(updateCompleteErrors)
	log.Println("No errors detected while sending complete pkt 🎉")
	log.Println("Everything is ok don't worry! Be happy 🎉🎉🎉")
	os.Exit(0)
}

type Serializer = func(clientIP udpack.IPString) ([][]byte, error)

func (updater *Updater) sendToEachClient(serialize Serializer) <-chan error {
	var wg sync.WaitGroup
	clientCount := len(updater.clients)
	errs := make(chan error, clientCount)
	for _, clientIP := range updater.clients {
		wg.Add(1)
		go updater.serializeAndSend(clientIP, serialize, errs, &wg)
	}
	wg.Wait()
	close(errs)
	return errs
}

func (updater *Updater) serializeAndSend(clientIP udpack.IPString, serialize Serializer, errs chan error, wg *sync.WaitGroup) {
	defer wg.Done()

	pkts, err := serialize(clientIP)
	if err != nil {
		errs <- err
		return
	}
	for _, pkt := range pkts {
		// @incomplete needs refactor
		udpAddr := &net.UDPAddr{
			IP:   net.ParseIP(string(clientIP)),
			Port: 8765,
			Zone: "",
		}
		err = updater.conn.WriteTo(pkt, udpAddr)
		if err != nil {
			errs <- err
			return
		}
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
	return cell.LinkOptions == other.LinkOptions &&
		cell.TimeSlot == other.TimeSlot &&
		cell.Channel == other.Channel
}

type PktType int

const (
	PktTypeUpdateRequest = iota
	PktTypeUpdateConfirmation
)

type UpdateRequest struct {
	NeighborAddr net.IP
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

type Schedule map[udpack.IPString]map[udpack.IPString][]Cell

func NewSchedule() Schedule {
	return make(Schedule)
}

func (schedulePtr *Schedule) AddCell(nodeAddr udpack.IPString, neighborAddr udpack.IPString, cell *Cell) {
	schedule := *schedulePtr
	mapCells, in := schedule[nodeAddr]
	if !in {
		mapCells = make(map[udpack.IPString][]Cell)
		schedule[nodeAddr] = mapCells
	}
	cells, in := mapCells[neighborAddr]
	if !in {
		cells = []Cell{}
		mapCells[neighborAddr] = cells
	}
	mapCells[neighborAddr] = append(cells, *cell)
}

func (schedulePtr Schedule) Serialize(clientIP udpack.IPString) ([][]byte, error) {
	clientSchedule, in := schedulePtr[clientIP]
	if !in {
		return nil, errors.New(fmt.Sprintf("the udpack ip address (%s) doesn't have any schedule associated with it", clientIP))
	}

	pkts := make([][]byte, 0)
	for neighborAddr, neighborCells := range clientSchedule {
		log.Printf("NeighborAddr : %+v\n", neighborAddr)
		for i := 0; i < len(neighborCells); i += ScheduleUpdaterPktMaxCells {
			pkt := UpdateRequest{
				NeighborAddr: net.ParseIP(string(neighborAddr)),
				Cells:        neighborCells[i:min(i+ScheduleUpdaterPktMaxCells, len(neighborCells))],
			}
			pkts = append(pkts, pkt.Encode())
		}
	}

	log.Println("Number of pkts to send: ", len(pkts))
	return pkts, nil
}

func (schedulePtr Schedule) IsCellUsed(nodeAddr udpack.IPString, neighborAddr udpack.IPString, cell *Cell) bool {
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
