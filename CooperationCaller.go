package stormi

import (
	"encoding/json"
	"math/rand"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/google/uuid"
)

type CooperationCaller struct {
	rp             *RedisProxy
	coprotocolId   string
	uuid           string
	concurrency    int
	receivebuffer  []reciveBufferStruct
	slots          chan int
	handlermapmap  map[int]map[string]time.Duration
	handlermaplist map[int][]string
	handlercount   map[int]int
}

type reciveBufferStruct struct {
	data chan []byte
	uuid string
}

func (cop *CooperationProxy) NewCaller() *CooperationCaller {
	copcl := CooperationCaller{}
	copcl.rp = cop.cp.rp
	copcl.concurrency = 100
	copcl.uuid = uuid.NewString()
	copcl.coprotocolId = cop.uuid
	copcl.handlermapmap = make(map[int]map[string]time.Duration)
	return &copcl
}

func (copcl *CooperationCaller) SetConcurrency(concurrency int) {
	copcl.concurrency = concurrency
}

func (copcl *CooperationCaller) initHanderMapMap() {
	if copcl.receivebuffer == nil {
		copcl.receivebuffer = make([]reciveBufferStruct, copcl.concurrency)
		for index := range copcl.receivebuffer {
			copcl.receivebuffer[index].data = make(chan []byte, 1)

		}
		copcl.slots = make(chan int, copcl.concurrency)
		for i := 0; i < copcl.concurrency; i++ {
			copcl.slots <- i
		}
		pubsub1 := copcl.rp.GetPubSub(copcl.coprotocolId)
		go func() {
			for {
				copcl.rp.Notify(copcl.coprotocolId, hi)
				time.Sleep(10 * time.Second)
			}
		}()
		go func() {
			copcl.rp.Subscribe(pubsub1, 0, func(msg string) int {
				if msg != hi {
					parts := strings.Split(msg, "@")
					if len(parts) == 3 {
						mtd, _ := strconv.Atoi(parts[1])
						d, err := time.ParseDuration(parts[2])
						if err == nil {
							if copcl.handlermapmap[mtd] == nil {
								copcl.handlermapmap[mtd] = make(map[string]time.Duration)
							}
							copcl.handlermapmap[mtd][parts[0]] = d
						}
					}
				}
				return 0
			})
		}()
		pubsub2 := copcl.rp.GetPubSub(copcl.uuid)
		go func() {
			copcl.rp.Subscribe(pubsub2, 0, func(msg string) int {
				copdto := cooperationDto{}
				json.Unmarshal([]byte(msg), &copdto)
				if copcl.receivebuffer[copdto.Slot].uuid == copdto.CallerUUID {
					copcl.receivebuffer[copdto.Slot].data <- copdto.Data
				}
				return 0
			})
		}()
	}
}

func (copcl *CooperationCaller) initHandlerMapListAndCount(method int) {
	if copcl.handlercount[method] == 0 {
		l := []string{}
		for uuid := range copcl.handlermapmap[method] {
			l = append(l, uuid)
		}
		handlermaplistlock.Lock()
		copcl.handlermaplist[method] = l
		copcl.handlercount[method] = len(l)
		handlermaplistlock.Unlock()
	}
}

var handlermaplistlock sync.Mutex

func (copcl *CooperationCaller) removeOneInHandlerMapListAndCount(method int, uuid string) {
	if uuid == "" {
		return
	}
	for index, u := range copcl.handlermaplist[method] {
		if u == uuid {
			handlermaplistlock.Lock()
			copcl.handlermaplist[method][index] = ""
			copcl.handlercount[method]--
			handlermaplistlock.Unlock()
			break
		}
	}
}

func (copcl *CooperationCaller) choose(method int) string {
	copcl.initHanderMapMap()
	copcl.initHandlerMapListAndCount(method)
	for {
		if copcl.handlercount[method] == 0 {
			return ""
		} else {
			index := rand.Intn(copcl.handlercount[method])
			if copcl.handlermaplist[method][index] != "" {
				return copcl.handlermaplist[method][index]
			}
		}
	}
}

func (copcl *CooperationCaller) Call(method int, send any, receive any) {
	slot := <-copcl.slots
	copcl.receivebuffer[slot].uuid = uuid.NewString()
	if len(copcl.receivebuffer[slot].data) == 1 {
		<-copcl.receivebuffer[slot].data
	}
	senddata, err := json.Marshal(send)
	copdto := cooperationDto{}
	if err == nil {
		copdto.Data = senddata
		copdto.CallerUUID = copcl.receivebuffer[slot].uuid
		copdto.CallerChannel = copcl.uuid
		copdto.Slot = slot
	}
	hid := copcl.choose(method)
	var receivedate []byte
	if hid != "" {
		data := []byte{}
		json.Unmarshal(data, &copdto)
		copcl.rp.Notify(hid, string(data))
		t := time.NewTicker(3*time.Second + copcl.handlermapmap[method][hid])
		select {
		case <-t.C:
		case receivedate = <-copcl.receivebuffer[slot].data:
		}
	}
	json.Unmarshal(receivedate, receive)
	if receivedate == nil {
		copcl.removeOneInHandlerMapListAndCount(method, hid)
	}
	copcl.receivebuffer[slot].uuid = ""
	copcl.slots <- slot
}
