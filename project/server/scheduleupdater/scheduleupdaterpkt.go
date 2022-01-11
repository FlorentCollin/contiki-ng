package scheduleupdater

import (
	"bytes"
	"encoding/binary"
	"errors"
	"fmt"
	"log"
	"net"
	"sync"

	"github.com/plgd-dev/go-coap/v2/message"
	"github.com/plgd-dev/go-coap/v2/message/codes"
	"github.com/plgd-dev/go-coap/v2/mux"
)

const ScheduleUpdaterPktMaxCells = 20

type Updater struct {
	clientsStates map[string]*ClientState
}

func NewUpdaterState() Updater {
	return Updater{
		clientsStates: make(map[string]*ClientState),
	}
}

func (s *Updater) AddClient(client mux.Client, token message.Token) {
	s.clientsStates[client.RemoteAddr().String()] = &ClientState{client, token, 1}
}

func (s *Updater) UpdateClients(schedule *Schedule) {
	log.Println("Called")
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

	//Update complete
    updateCompleteErrors := s.sendToEachClient(func(ip string) ([][]byte, error) {
        buffer := new(bytes.Buffer)
        pkt := UpdateCompletePkt{}
        err := pkt.Encode(buffer)
        if err != nil {
            return nil, err
        }
        return [][]byte{buffer.Bytes()}, nil
    })
    panicIfErrors(updateCompleteErrors)
    log.Println("No errors detected while sending complete pkt 🎉")
}

type Serializer = func(ip string) ([][]byte, error)

func (s *Updater) sendToEachClient(serialize Serializer) <-chan error {
	var wg sync.WaitGroup
	clientCount := len(s.clientsStates)
	errs := make(chan error, clientCount)
	for clientIp, clientState := range s.clientsStates {
		wg.Add(1)
		go serializeAndSend(clientIp, clientState, serialize, errs, &wg)
	}
	wg.Wait()
	close(errs)
	return errs
}

func serializeAndSend(clientIp string, clientState *ClientState, serialize Serializer, errs chan error, wg *sync.WaitGroup) {
	defer wg.Done()

	messages, err := serialize(clientIp)
	if err != nil {
		errs <- err
		return
	}
    for _, message := range(messages) {
        err = SendToClient(clientState.client, clientState.token, clientState.nextObs(), &message)
        if err != nil {
            errs <- err
            return
        }
    }
}

type NeighborAddr [4]uint16

type Cell struct {
	LinkOptions  uint8
	TimeSlot     uint16
	Channel      uint16
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
    err := binary.Write(buffer, binary.LittleEndian, []byte("   ")) // Fix bug
	if err != nil {
		return err
	}
	err = binary.Write(buffer, binary.LittleEndian, uint8(pkt.Type()))
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
    err := binary.Write(buffer, binary.LittleEndian, []byte("   ")); // FIX some bug
    if err != nil {
        return err;
    }
	return binary.Write(buffer, binary.LittleEndian, uint8(pkt.Type()))
}

type ClientState struct {
	client  mux.Client
	token   message.Token
	lastObs uint32
}

func (c *ClientState) nextObs() uint32 {
	c.lastObs = c.lastObs + 1
	return c.lastObs
}

type Schedule map[string]map[NeighborAddr][]Cell

func NewSchedule() Schedule {
	return make(Schedule)
}

func (schedulePtr *Schedule) AddCell(node net.Addr, neihborAddr NeighborAddr, cell *Cell) {
	schedule := *schedulePtr
	map_cells, in := schedule[node.String()]
	if !in {
		map_cells = make(map[NeighborAddr][]Cell)
		schedule[node.String()] = map_cells
	}
    cells, in := map_cells[neihborAddr]
    if !in {
        cells = []Cell{}
        map_cells[neihborAddr] = cells
    }
	map_cells[neihborAddr] = append(cells, *cell)
}

func (schedulePtr *Schedule) Serialize(ip string) ([][]byte, error) {
	clientSchedule, in := (*schedulePtr)[ip]
	if !in {
		return nil, errors.New("the client ip address doesn't have any schedule associated with it")
	}

    messages := make([][]byte, 0)
	for neighborAddr, neighborCells := range clientSchedule {
        log.Printf("NeighborAddr : %+v\n", neighborAddr);
        for i := 0; i < len(neighborCells); i += ScheduleUpdaterPktMaxCells {
            pkt := UpdatePkt{
                NeighborAddr: neighborAddr,
                Cells: neighborCells[i:min(i + ScheduleUpdaterPktMaxCells, len(neighborCells))],
            }
            buffer := new(bytes.Buffer)
            err := pkt.Encode(buffer)
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

func SendToClient(cc mux.Client, token []byte, obs uint32, messageBody *[]byte) error {
	m := message.Message{
		Code:    codes.Content,
		Token:   token,
		Context: cc.Context(),
		Body:    bytes.NewReader(*messageBody),
	}
	var opts message.Options
	var buf []byte
	opts, n, err := opts.SetContentFormat(buf, message.TextPlain)
	if err == message.ErrTooSmall {
		buf = append(buf, make([]byte, n)...)
		opts, n, err = opts.SetContentFormat(buf, message.TextPlain)
	}
	if err != nil {
		return fmt.Errorf("cannot set content format to response: %w", err)
	}
	opts, n, err = opts.SetObserve(buf, obs)
	if err == message.ErrTooSmall {
		buf = append(buf, make([]byte, n)...)
		opts, n, err = opts.SetObserve(buf, obs)
	}
	if err != nil {
		return fmt.Errorf("cannot set options to response: %w", err)
	}
	m.Options = opts
	log.Println("Sending to client ", cc.RemoteAddr().String(), " with obs ", obs)
	return cc.WriteMessage(&m)
}

func min(x, y int) int {
    if x <= y {
        return x
    }
    return y
}
