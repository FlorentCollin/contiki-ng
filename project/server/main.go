package main

import (
	"bytes"
	"encoding/binary"
	"fmt"
	"io"
	"log"
	"time"

	"github.com/plgd-dev/go-coap/v2"
	"github.com/plgd-dev/go-coap/v2/message"
	"github.com/plgd-dev/go-coap/v2/message/codes"
	"github.com/plgd-dev/go-coap/v2/mux"
)

type CellID struct {
	LinkOptions uint8
	TimeSlot    uint16
	Channel     uint16
}

type ScheduleUpdaterPkt struct {
	NeighborAddr [4]uint16
	CellIDsLen   uint8
	CellIDs      []CellID
}

func writeOrPanic(w io.Writer, order binary.ByteOrder, data interface{}) {
	err := binary.Write(w, order, data)
	if err != nil {
		log.Panicf("Error while writing to the binary buffer : %v", err)
	}
}

func (pkt *ScheduleUpdaterPkt) Encode(buffer *bytes.Buffer) {
	writeOrPanic(buffer, binary.LittleEndian, [8]uint8{1, 2, 3, 4, 5, 6, 7, 8})
	writeOrPanic(buffer, binary.LittleEndian, pkt.CellIDsLen)
	for i := 0; uint8(i) < pkt.CellIDsLen; i++ {
		writeOrPanic(buffer, binary.LittleEndian, pkt.CellIDs[i].LinkOptions)
		writeOrPanic(buffer, binary.LittleEndian, pkt.CellIDs[i].TimeSlot)
		writeOrPanic(buffer, binary.LittleEndian, pkt.CellIDs[i].Channel)
	}
}

func loggingMiddleware(next mux.Handler) mux.Handler {
	return mux.HandlerFunc(func(w mux.ResponseWriter, r *mux.Message) {
		log.Printf("Got a message path=%v: %+v from %v\n", getPath(r.Options), r, w.Client().RemoteAddr())
		next.ServeCOAP(w, r)
	})
}

func getPath(opts message.Options) string {
	path, err := opts.Path()
	if err != nil {
		log.Printf("cannot get path: %v", err)
		return ""
	}
	return path
}

func periodicTransmitter(cc mux.Client, token []byte) {
	for obs := int64(2); ; obs++ {
		pkt := ScheduleUpdaterPkt{
			NeighborAddr: [4]uint16{0, 0, 0, 0},
			CellIDsLen:   1,
			CellIDs: []CellID{
				{LinkOptions: 2, TimeSlot: 4, Channel: 5},
			},
		}
		buffer := new(bytes.Buffer)
		pkt.Encode(buffer)

		err := sendResponse(cc, token, obs, buffer.Bytes())
		if err != nil {
			log.Panic("Error in transmitter, stopping: %v", err)
			return
		}
		time.Sleep(time.Second)
	}
}

func sendResponse(cc mux.Client, token []byte, obs int64, messageBody []byte) error {
	m := message.Message{
		Code:    codes.Content,
		Token:   token,
		Context: cc.Context(),
		Body:    bytes.NewReader(messageBody),
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
	if obs >= 0 {
		opts, n, err = opts.SetObserve(buf, uint32(obs))
		if err == message.ErrTooSmall {
			buf = append(buf, make([]byte, n)...)
			opts, n, err = opts.SetObserve(buf, uint32(obs))
		}
		if err != nil {
			return fmt.Errorf("cannot set options to response: %w", err)
		}
	}
	m.Options = opts
	log.Printf("Message: %+v\n", m)
	log.Printf("messageBody: %s", messageBody)
	return cc.WriteMessage(&m)
}

func handleSchedule(w mux.ResponseWriter, r *mux.Message) {
	obs, err := r.Options.Observe()
	log.Printf("Message %+v\n", r)
	log.Printf("obs %+v\n", obs)
	if err != nil {
		log.Panicf("Error while retrieving observer options %v\n", err)
	}
	if r.Code == codes.GET && err == nil && obs == 0 {
		log.Println("Start sending periodic message")
		go periodicTransmitter(w.Client(), r.Token)
	}
}

func main() {
	r := mux.NewRouter()
	r.Use(loggingMiddleware)
	err := r.Handle("/schedule", mux.HandlerFunc(handleSchedule))
	if err != nil {
		log.Panicf("Error while trying to handle the new route %v\n", err)
	}

	log.Printf("Starting CoAP Server on port 5683")
	log.Fatal(coap.ListenAndServe("udp", ":5683", r))
}
