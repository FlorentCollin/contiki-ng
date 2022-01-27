package scheduleupdater

import (
	"coap-server/client"
	"coap-server/utils"
	"errors"
	"log"
	"sync"
	"sync/atomic"
)

const ScheduleUpdaterPktMaxCells = 11

type Updater struct {
	PackagesSentCount uint32 // atomic value
	clientsStates     map[string]*client.State
}

func NewUpdater(clientsStates []*client.State) Updater {
	clientsMap := make(map[string]*client.State)
	for _, clientState := range clientsStates {
		clientsMap[clientState.Addr.IP.String()] = clientState
	}
	return Updater{
		PackagesSentCount: 0,
		clientsStates:     clientsMap,
	}
}

func (s *Updater) UpdateClients(schedule *Schedule) {
	// helper function to check if any error and panic if any
	panicIfErrors := func(errs <-chan error) {
		for err := range errs {
			if err != nil {
				log.Panic(err.Error())
			}
		}
	}

	// New schedule update
	scheduleUpdateErrors := s.sendToEachClient(schedule.Serialize)
	panicIfErrors(scheduleUpdateErrors)
	log.Println("No errors detected while sending the new schedule 🎉")

	// Update complete
	updateCompleteErrors := s.sendToEachClient(func(clientState *client.State) ([]client.PktEncoder, error) {
		updateCompletePkt := UpdateCompletePkt{}
		return []client.PktEncoder{&updateCompletePkt}, nil
	})
	panicIfErrors(updateCompleteErrors)
	log.Println("No errors detected while sending complete pkt 🎉")
}

type Serializer = func(clientState *client.State) ([]client.PktEncoder, error)

func (s *Updater) sendToEachClient(serialize Serializer) <-chan error {
	var wg sync.WaitGroup
	clientCount := len(s.clientsStates)
	errs := make(chan error, clientCount)
	for _, clientState := range s.clientsStates {
		wg.Add(1)
		go s.serializeAndSend(clientState, serialize, errs, &wg)
	}
	wg.Wait()
	close(errs)
	return errs
}

func (s *Updater) serializeAndSend(clientState *client.State, serialize Serializer, errs chan error, wg *sync.WaitGroup) {
	defer wg.Done()

	pkts, err := serialize(clientState)
	if err != nil {
		errs <- err
		return
	}
	for _, pkt := range pkts {
		atomic.AddUint32(&s.PackagesSentCount, 1)
		err = clientState.Send(pkt)
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

func (pkt *UpdatePkt) Encode(buffer []byte) []byte {
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

func (pkt *UpdateCompletePkt) Encode(buffer []byte) []byte {
	return append(buffer, uint8(pkt.Type()))
}

type Schedule map[string]map[NeighborAddr][]Cell

func NewSchedule() Schedule {
	return make(Schedule)
}

func (schedulePtr *Schedule) AddCell(nodeAddr string, neihborAddr NeighborAddr, cell *Cell) {
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

func (schedulePtr *Schedule) Serialize(clientState *client.State) ([]client.PktEncoder, error) {
	ip := clientState.Addr.IP.String()
	clientSchedule, in := (*schedulePtr)[ip]
	if !in {
		return nil, errors.New("the client ip address doesn't have any schedule associated with it")
	}

	pkts := make([]client.PktEncoder, 0)
	for neighborAddr, neighborCells := range clientSchedule {
		log.Printf("NeighborAddr : %+v\n", neighborAddr)
		for i := 0; i < len(neighborCells); i += ScheduleUpdaterPktMaxCells {
			pkt := UpdatePkt{
				NeighborAddr: neighborAddr,
				Cells:        neighborCells[i:min(i+ScheduleUpdaterPktMaxCells, len(neighborCells))],
			}
			pkts = append(pkts, &pkt)
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
