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
	rp                    *RedisProxy
	coprotocolId          string
	uuid                  string
	concurrency           int
	receivebuffer         []reciveBufferStruct
	slots                 chan int
	handlermapmap         map[int]map[string]time.Time
	handlermaplist        map[int][]string
	timeout               time.Duration
	concurrentmaplistlock sync.Mutex
	concurrentmapmaplock  sync.Mutex
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
	copcl.timeout = 5 * time.Second
	copcl.handlermapmap = make(map[int]map[string]time.Time)
	return &copcl
}

func (copcl *CooperationCaller) SetConcurrency(concurrency int) {
	copcl.concurrency = concurrency
}

func (copcl *CooperationCaller) SetTimeout(timeout time.Duration) {
	copcl.timeout = timeout
}

func (copcl *CooperationCaller) init() {
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
			copcl.rp.Subscribe(pubsub1, 0, func(msg string) int {
				if msg != hi {
					parts := strings.Split(msg, "@")
					if len(parts) == 2 {
						mtd, _ := strconv.Atoi(parts[1])
						copcl.concurrentmaplistlock.Lock()
						if len(copcl.handlermaplist[mtd]) == 0 {
							copcl.concurrentmaplistlock.Unlock()
							copcl.refreshHandlerMapList()
						} else {
							copcl.concurrentmaplistlock.Unlock()
						}
						copcl.concurrentmapmaplock.Lock()
						if copcl.handlermapmap[mtd] == nil {
							copcl.handlermapmap[mtd] = make(map[string]time.Time)
						}
						copcl.handlermapmap[mtd][parts[0]] = time.Now()
						copcl.concurrentmapmaplock.Unlock()
					}

				}
				return 0
			})
		}()
		go func() {
			for {
				copcl.refreshHandlerMapList()
				time.Sleep(1 * time.Second)
			}
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

func (copcl *CooperationCaller) refreshHandlerMapList() {
	for method := range copcl.handlermapmap {
		if copcl.handlermaplist == nil {
			copcl.handlermaplist = make(map[int][]string)
		}
		if copcl.handlermaplist[method] == nil {
			copcl.handlermaplist[method] = []string{}
		}
		for key, t := range copcl.handlermapmap[method] {
			if time.Since(t) > 3*time.Second {
				copcl.concurrentmapmaplock.Lock()
				delete(copcl.handlermapmap[method], key)
				copcl.concurrentmapmaplock.Unlock()
				for index, v := range copcl.handlermaplist[method] {
					if key == v {
						copcl.concurrentmaplistlock.Lock()
						copcl.removeOneInHandlerMapList(method, index)
						copcl.concurrentmaplistlock.Unlock()
						break
					}
				}
			}
			if time.Since(t) < 1500*time.Millisecond {
				var exist bool
				for _, v := range copcl.handlermaplist[method] {
					if key == v {
						exist = true
						break
					}
				}
				if !exist {
					copcl.concurrentmaplistlock.Lock()
					copcl.handlermaplist[method] = append(copcl.handlermaplist[method], key)
					copcl.concurrentmaplistlock.Unlock()
				}
			}
		}
	}
}

func (copcl *CooperationCaller) removeOneInHandlerMapList(method int, index int) {
	if index == len(copcl.handlermaplist[method])-1 {
		copcl.handlermaplist[method] = copcl.handlermaplist[method][:index]
	} else {
		copcl.handlermaplist[method] = append(copcl.handlermaplist[method][:index], copcl.handlermaplist[method][index+1:]...)
	}
}

func (copcl *CooperationCaller) choose(method int) string {
	copcl.concurrentmaplistlock.Lock()
	if len(copcl.handlermaplist[method]) == 0 {
		copcl.rp.Notify(copcl.coprotocolId, hi)
		copcl.concurrentmaplistlock.Unlock()
		return ""
	} else {
		index := rand.Intn(len(copcl.handlermaplist[method]))
		copcl.concurrentmaplistlock.Unlock()
		return copcl.handlermaplist[method][index]
	}
}
func (copcl *CooperationCaller) chooseByTimeAndHandler(method int, timeremain time.Duration, handlerlist []string) string {
	if len(handlerlist) == len(copcl.handlermaplist[method]) {
		return ""
	}
	handlerChannelId := copcl.choose(method)
	retrynum := int(timeremain) / int(time.Second)
	ok := true
	for i := 0; i < retrynum; i++ {
		ok = true
		if handlerChannelId != "" {
			break
		}
		time.Sleep(1 * time.Second)
		handlerChannelId = copcl.choose(method)
		if i == retrynum-1 {
			return ""
		}
		for _, h := range handlerlist {
			if h == handlerChannelId {
				ok = false
				break
			}
		}
		if ok {
			break
		}
	}
	return handlerChannelId
}

func (copcl *CooperationCaller) Call(method int, send, receive any) {
	copcl.init()
	calledhandlerlist := []string{}
	now := time.Now()

	slot := <-copcl.slots
	copcl.receivebuffer[slot].uuid = uuid.NewString()
	if len(copcl.receivebuffer[slot].data) == 1 {
		<-copcl.receivebuffer[slot].data
	}
	senddata, err := json.Marshal(send)
	copdto := cooperationDto{}

	if err != nil {
		return
	}
	var receivedate []byte

	t := time.NewTicker(copcl.timeout)

	finish := false
	go func() {
		<-t.C
		finish = true
	}()
	for {
		if finish {
			break
		}

		timeremaining := copcl.timeout - time.Since(now)
		if timeremaining <= 0 {
			return
		}
		copdto.Data = senddata
		copdto.CallerUUID = copcl.receivebuffer[slot].uuid
		copdto.CallerChannel = copcl.uuid
		copdto.Slot = slot
		copdto.TimeRemaining = timeremaining
		tt := time.NewTicker(timeremaining)
		data, err := json.Marshal(copdto)
		if err != nil {
			return
		}
		handlerChannelId := copcl.chooseByTimeAndHandler(method, copcl.timeout, calledhandlerlist)
		if handlerChannelId == "" {
			time.Sleep(time.Duration(rand.Intn(1000) * int(time.Millisecond)))
			calledhandlerlist = []string{}
			continue
		} else {
			calledhandlerlist = append(calledhandlerlist, handlerChannelId)
		}
		copcl.rp.Notify(handlerChannelId, string(data))
		select {
		case <-tt.C:
		case receivedate = <-copcl.receivebuffer[slot].data:
			tt.Stop()
		}
		if receivedate != nil {
			break
		}
	}
	t.Stop()
	json.Unmarshal(receivedate, receive)

	copcl.receivebuffer[slot].uuid = ""
	copcl.slots <- slot
}
