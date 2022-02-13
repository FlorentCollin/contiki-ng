package scheduleupdater

import (
	"coap-server/udpack"
	"coap-server/utils"
	"errors"
	"log"
	"net"
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
		updateCompletePkt := UpdateCompletePkt{}
		return [][]byte{updateCompletePkt.Encode()}, nil
	})
	panicIfErrors(updateCompleteErrors)
	log.Println("No errors detected while sending complete pkt 🎉")
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

type NeighborAddr [4]uint16

type Cell struct {
	LinkOptions uint8
	TimeSlot    uint16
	Channel     uint16
}

type PktType int

const (
	PktTypeUpdate = iota
	PktTypeUpdateComplete
)

type UpdatePkt struct {
	NeighborAddr [4]uint16
	Cells        []Cell
}

func (pkt *UpdatePkt) Type() PktType {
	return PktTypeUpdate
}

func (pkt *UpdatePkt) Encode() []byte {
	buffer := make([]byte, 0)
	buffer = append(buffer, uint8(pkt.Type()))
	// Appending the neighborAddr uint16 to the buffer
	for _, v := range pkt.NeighborAddr {
		buffer = utils.AppendBigEndianUint16(buffer, v)
	}
	buffer = append(buffer, uint8(len(pkt.Cells)))
	for i := 0; i < len(pkt.Cells); i++ {
		buffer = append(buffer, pkt.Cells[i].LinkOptions)
		buffer = utils.AppendLittleEndianUint16(buffer, pkt.Cells[i].TimeSlot)
		buffer = utils.AppendLittleEndianUint16(buffer, pkt.Cells[i].Channel)
	}
	return buffer
}

type UpdateCompletePkt struct{}

func (pkt *UpdateCompletePkt) Type() PktType {
	return PktTypeUpdateComplete
}

func (pkt *UpdateCompletePkt) Encode() []byte {
	return []byte{uint8(pkt.Type())}
}

type Schedule map[udpack.IPString]map[NeighborAddr][]Cell

func NewSchedule() Schedule {
	return make(Schedule)
}

func (schedulePtr *Schedule) AddCell(nodeAddr udpack.IPString, neihborAddr NeighborAddr, cell *Cell) {
	schedule := *schedulePtr
	mapCells, in := schedule[nodeAddr]
	if !in {
		mapCells = make(map[NeighborAddr][]Cell)
		schedule[nodeAddr] = mapCells
	}
	cells, in := mapCells[neihborAddr]
	if !in {
		cells = []Cell{}
		mapCells[neihborAddr] = cells
	}
	mapCells[neihborAddr] = append(cells, *cell)
}

func (schedulePtr Schedule) Serialize(clientIP udpack.IPString) ([][]byte, error) {
	clientSchedule, in := schedulePtr[clientIP]
	if !in {
		return nil, errors.New("the udpack ip address doesn't have any schedule associated with it")
	}

	pkts := make([][]byte, 0)
	for neighborAddr, neighborCells := range clientSchedule {
		log.Printf("NeighborAddr : %+v\n", neighborAddr)
		for i := 0; i < len(neighborCells); i += ScheduleUpdaterPktMaxCells {
			pkt := UpdatePkt{
				NeighborAddr: neighborAddr,
				Cells:        neighborCells[i:min(i+ScheduleUpdaterPktMaxCells, len(neighborCells))],
			}
			pkts = append(pkts, pkt.Encode())
		}
	}

	log.Println("Number of pkts to send: ", len(pkts))
	return pkts, nil
}

// ----- INTERNAL ----

func min(x, y int) int {
	if x <= y {
		return x
	}
	return y
}
