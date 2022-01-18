package scheduleupdater

import (
	"bytes"
	"coap-server/client"
	"encoding/binary"
	"errors"
	"fmt"
	"log"
	"sync"
	"sync/atomic"
)

const ScheduleUpdaterPktMaxCells = 20

type Updater struct {
	PackagesSentCount uint32 // atomic value
	clientsStates     map[string]*client.ClientState
}

func NewUpdater(clientsStates []*client.ClientState) Updater {
	clientsMap := make(map[string]*client.ClientState)
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
				log.Panicf(err.Error())
			}
		}
	}

	// New schedule update
	scheduleUpdateErrors := s.sendToEachClient(schedule.Serialize)
	panicIfErrors(scheduleUpdateErrors)
	log.Println("No errors detected while sending the new schedule 🎉")

	// Update complete
	updateCompleteErrors := s.sendToEachClient(func(clientState *client.ClientState) ([][]byte, error) {
		buffer := new(bytes.Buffer)
		pkt := UpdateCompletePkt{}
		err := clientState.EncodePktNumber(buffer)
		if err != nil {
			return nil, err
		}
		err = pkt.Encode(buffer)
		if err != nil {
			return nil, err
		}
		return [][]byte{buffer.Bytes()}, nil
	})
	panicIfErrors(updateCompleteErrors)
	log.Println("No errors detected while sending complete pkt 🎉")
}

type Serializer = func(clientState *client.ClientState) ([][]byte, error)

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

func (s *Updater) serializeAndSend(clientState *client.ClientState, serialize Serializer, errs chan error, wg *sync.WaitGroup) {
	defer wg.Done()

	messages, err := serialize(clientState)
	if err != nil {
		errs <- err
		return
	}
	for _, message := range messages {
		atomic.AddUint32(&s.PackagesSentCount, 1)
		err = clientState.Send(message)
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

type Pkt interface {
	Type() PktType
	Encode(buffer *bytes.Buffer) error
}

type UpdatePkt struct {
	NeighborAddr [4]uint16
	Cells        []Cell
}

func (pkt *UpdatePkt) Type() PktType {
	return PktTypeUpdate
}

func (pkt *UpdatePkt) Encode(buffer *bytes.Buffer) error {
	err := binary.Write(buffer, binary.LittleEndian, uint8(pkt.Type()))
	if err != nil {
		return err
	}
	err = binary.Write(buffer, binary.LittleEndian, pkt.NeighborAddr)
	if err != nil {
		return err
	}
	err = binary.Write(buffer, binary.LittleEndian, uint8(len(pkt.Cells)))
	if err != nil {
		return err
	}
	for i := 0; i < len(pkt.Cells); i++ {
		err = binary.Write(buffer, binary.LittleEndian, pkt.Cells[i].LinkOptions)
		if err != nil {
			return err
		}
		err = binary.Write(buffer, binary.LittleEndian, pkt.Cells[i].TimeSlot)
		if err != nil {
			return err
		}
		err = binary.Write(buffer, binary.LittleEndian, pkt.Cells[i].Channel)
		if err != nil {
			return err
		}
	}
	return nil
}

type UpdateCompletePkt struct{}

func (pkt *UpdateCompletePkt) Type() PktType {
	return PktTypeUpdateComplete
}

func (pkt *UpdateCompletePkt) Encode(buffer *bytes.Buffer) error {
	return binary.Write(buffer, binary.LittleEndian, uint8(pkt.Type()))
}

type Schedule map[string]map[NeighborAddr][]Cell

func NewSchedule() Schedule {
	return make(Schedule)
}

func (schedulePtr *Schedule) AddCell(nodeAddr string, neihborAddr NeighborAddr, cell *Cell) {
	schedule := *schedulePtr
	map_cells, in := schedule[nodeAddr]
	if !in {
		map_cells = make(map[NeighborAddr][]Cell)
		schedule[nodeAddr] = map_cells
	}
	cells, in := map_cells[neihborAddr]
	if !in {
		cells = []Cell{}
		map_cells[neihborAddr] = cells
	}
	map_cells[neihborAddr] = append(cells, *cell)
}

func (schedulePtr *Schedule) Serialize(clientState *client.ClientState) ([][]byte, error) {
	ip := clientState.Addr.IP.String()
	clientSchedule, in := (*schedulePtr)[ip]
	if !in {
		return nil, errors.New("the client ip address doesn't have any schedule associated with it")
	}

	messages := make([][]byte, 0)
	for neighborAddr, neighborCells := range clientSchedule {
		log.Printf("NeighborAddr : %+v\n", neighborAddr)
		for i := 0; i < len(neighborCells); i += ScheduleUpdaterPktMaxCells {
			pkt := UpdatePkt{
				NeighborAddr: neighborAddr,
				Cells:        neighborCells[i:min(i+ScheduleUpdaterPktMaxCells, len(neighborCells))],
			}
			buffer := new(bytes.Buffer)
			err := clientState.EncodePktNumber(buffer)
			if err != nil {
				return nil, fmt.Errorf("Error while encoding the pkt into the buffer, stopping %+v\n", err)
			}
			err = pkt.Encode(buffer)
			if err != nil {
				return nil, fmt.Errorf("Error while encoding the pkt into the buffer, stopping %+v\n", err)
			}
			messages = append(messages, buffer.Bytes())
		}
	}

	log.Println("Number of messages to send: ", len(messages))
	return messages, nil
}

// ----- INTERNAL ----

func min(x, y int) int {
	if x <= y {
		return x
	}
	return y
}
