package stats

import (
	"encoding/json"
	"log"
	"os"
	"scheduleupdater-server/addrtranslation"
	"scheduleupdater-server/utils"
	"strconv"
	"sync"
	"time"
)

type IncDict struct {
	IPMap map[addrtranslation.IPString]int `json:"IPMap,omitempty"`
	lock  sync.RWMutex
}

func NewIncDict() IncDict {
	return IncDict{
		IPMap: make(map[addrtranslation.IPString]int),
		lock:  sync.RWMutex{},
	}
}

func (d *IncDict) Increment(ip addrtranslation.IPString) {
	d.lock.Lock()
	defer d.lock.Unlock()
	if _, in := d.IPMap[ip]; !in {
		d.IPMap[ip] = 1
		return
	}
	d.IPMap[ip]++
}

type Stats struct {
	Nsent               IncDict   `json:"nsent,omitempty"`
	Nreceived           IncDict   `json:"nreceived,omitempty"`
	Timeouts            IncDict   `json:"timeouts,omitempty"`
	ProtocolSent        IncDict   `json:"protocolSent,omitempty"`
	ProtocolReceived    IncDict   `json:"protocolReceived,omitempty"`
	ScheduleUpdateStart time.Time `json:"scheduleUpdateStart,omitempty"`
	ScheduleUpdateEnd   time.Time `json:"scheduleUpdateEnd,omitempty"`
	Nclients            uint      `json:"nclients,omitempty"`
}

var SimulationStats = Stats{
	Nreceived:        NewIncDict(),
	Nsent:            NewIncDict(),
	Timeouts:         NewIncDict(),
	ProtocolSent:     NewIncDict(),
	ProtocolReceived: NewIncDict(),
	Nclients:         0,
}

func (stats *Stats) WriteToFile(prefix string) {
	filename := prefix + strconv.FormatInt(time.Now().Unix(), 10) + ".json"
	utils.Log.Println("Writing stats to file.")
	statsJson, err := json.MarshalIndent(stats, "", "    ")
	if err != nil {
		log.Panicln(err)
	}
	err = os.WriteFile(filename, statsJson, 0644)
	if err != nil {
		log.Panicln(err)
	}
}
