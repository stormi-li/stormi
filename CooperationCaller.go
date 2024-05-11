package stormi

import (
	"encoding/json"
	"math/rand"
	"strconv"
	"strings"
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
				if msg == full {
					copcl.receivebuffer[copdto.Slot].data <- copdto.Data
				}
				json.Unmarshal([]byte(msg), &copdto)
				if copcl.receivebuffer[copdto.Slot].uuid == copdto.CallerUUID {
					copcl.receivebuffer[copdto.Slot].data <- copdto.Data
				}
				return 0
			})
		}()
	}
}

func (copcl *CooperationCaller) initHandlerMapList(method int) {
	if copcl.handlermaplist == nil {
		copcl.handlermaplist = make(map[int][]string)
	}
	if len(copcl.handlermaplist[method]) == 0 {
		l := []string{}
		for uuid := range copcl.handlermapmap[method] {
			l = append(l, uuid)
		}
		if copcl.handlermaplist[method] == nil {
			copcl.handlermaplist[method] = []string{}
		}
		copcl.handlermaplist[method] = l
	}
}

func (copcl *CooperationCaller) removeOneInHandlerMapList(method int, uuid string) {
	if uuid == "" {
		return
	}
	for index, u := range copcl.handlermaplist[method] {
		delete(copcl.handlermapmap[method], uuid)
		if u == uuid {
			if index == len(copcl.handlermaplist[method])-1 {
				copcl.handlermaplist[method] = copcl.handlermaplist[method][:index]
			} else {
				copcl.handlermaplist[method] = append(copcl.handlermaplist[method][:index], copcl.handlermaplist[method][index+1:]...)
			}
			break
		}
	}
}

func (copcl *CooperationCaller) choose(method int) string {
	copcl.initHanderMapMap()
	copcl.initHandlerMapList(method)
	for {
		if len(copcl.handlermaplist[method]) == 0 {
			return ""
		} else {
			index := rand.Intn(len(copcl.handlermaplist[method]))
			if copcl.handlermaplist[method][index] != "" {
				return copcl.handlermaplist[method][index]
			}
		}
	}
}

func (copcl *CooperationCaller) Call(method int, send, receive any) {
	hid := copcl.choose(method)
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
	var receivedate []byte
	if hid != "" {
		data, _ := json.Marshal(copdto)
		copcl.rp.Notify(hid, string(data))
		t := time.NewTicker(3*time.Second + copcl.handlermapmap[method][hid])
		select {
		case <-t.C:
		case receivedate = <-copcl.receivebuffer[slot].data:
		}
	}
	json.Unmarshal(receivedate, receive)
	if receivedate == nil {
		copcl.removeOneInHandlerMapList(method, hid)
	}
	copcl.receivebuffer[slot].uuid = ""
	copcl.slots <- slot
}
