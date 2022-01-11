package main

import (
	"coap-server/scheduleupdater"
	"fmt"
	"log"
	"net"
	"time"

	"github.com/plgd-dev/go-coap/v2"
	"github.com/plgd-dev/go-coap/v2/message/codes"
	"github.com/plgd-dev/go-coap/v2/mux"
)

const nClients = 2;

func loggingMiddleware(next mux.Handler) mux.Handler {
	return mux.HandlerFunc(func(w mux.ResponseWriter, r *mux.Message) {
		path, err := r.Options.Path()
		if err != nil {
			log.Printf("Error while retrieving the path in the options: %+v", err)
		} else {
			log.Printf("Got a message path=%v: %+v from %v\n", path, r, w.Client().RemoteAddr())
		}
		next.ServeCOAP(w, r)
	})
}

func addCell(s *scheduleupdater.Schedule, addr net.Addr, timeslot uint16, channel uint16) {
    neighborAddr := [4]uint16{0xf80, 0x202, 0x2, 0x2}
	s.AddCell(addr,
		neighborAddr,
		&scheduleupdater.Cell{
			LinkOptions: 1,
			TimeSlot:    timeslot,
			Channel:     channel,
		})
}

func periodicSend(scheduleUpdater *scheduleupdater.Updater, schedule *scheduleupdater.Schedule) {
	for {
		go scheduleUpdater.UpdateClients(schedule)
		time.Sleep(5 * time.Second)
	}
}

func initializeSchedule(clientCount uint) scheduleupdater.Schedule {
	schedule := scheduleupdater.NewSchedule()
	addrs := make([]*net.UDPAddr, clientCount)
	for i := uint(2); i < clientCount+2; i++ {
		fmt.Printf("Adding fd00::20%d:%d:%d:%d\n", i, i, i, i)
		addrs[i-2] = &net.UDPAddr{
			IP:   net.ParseIP(fmt.Sprintf("fd00::20%d:%d:%d:%d", i, i, i, i)),
			Port: 5683,
			Zone: "",
		}
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

func main() {
	log.SetFlags(log.Ltime | log.Lshortfile)
	scheduleUpdater := scheduleupdater.NewUpdaterState()
	schedule := initializeSchedule(nClients)
	r := mux.NewRouter()
	r.Use(loggingMiddleware)
	clientCount := 0
	err := r.Handle("/schedule", mux.HandlerFunc(func(w mux.ResponseWriter, r *mux.Message) {
		// Handle the schedule
		obs, err := r.Options.Observe()
		log.Printf("Message %+v\n", r)
		log.Printf("obs %+v\n", obs)
		if err != nil {
			log.Panicf("Error while retrieving observer options %v\n", err)
		}
		if r.Code == codes.GET && err == nil && obs == 0 {
			log.Println("Adding new client ! ", w.Client().RemoteAddr())
			scheduleUpdater.AddClient(w.Client(), r.Token)
			clientCount++
			if clientCount >= len(schedule) {
				go scheduleUpdater.UpdateClients(&schedule)
			}
		}
	}))

	if err != nil {
		log.Panicf("Error while trying to handle the new route %v\n", err)
	}

	log.Printf("Starting CoAP Server on port 5683")
	log.Fatal(coap.ListenAndServe("udp", ":5683", r))

}
