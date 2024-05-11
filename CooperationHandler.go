package stormi

import (
	"encoding/json"
	"strconv"
	"time"

	"github.com/google/uuid"
)

type CooperationHandler struct {
	coprotocolId string
	concurrency  int
	buffersize   int
	rp           *RedisProxy
}

func (cop *CooperationProxy) NewHandler() *CooperationHandler {
	cophd := CooperationHandler{}
	cophd.coprotocolId = cop.uuid
	cophd.concurrency = 10
	cophd.buffersize = 1000
	cophd.rp = cop.cp.rp
	return &cophd
}

func (cophd *CooperationHandler) SetConcurrency(concurrency int) {
	cophd.concurrency = concurrency
}

func (cophd *CooperationHandler) SetBufferSize(size int) {
	cophd.buffersize = size
}

type cooperationDto struct {
	Data          []byte
	CallerChannel string
	Slot          int
	CallerUUID    string
}

func (cophd *CooperationHandler) Handle(method int, handler func(data []byte) any) {
	receivebuffer := make(chan cooperationDto, cophd.buffersize)
	sendbuffer := make(chan cooperationDto, 100)
	channelname := cophd.coprotocolId
	pubsub1 := cophd.rp.GetPubSub(channelname)
	cophdid := uuid.NewString()
	pubsub2 := cophd.rp.GetPubSub(cophdid)
	var timeconsume time.Duration
	mtd := strconv.Itoa(method)
	cophd.rp.Notify(channelname, cophdid+"@"+mtd+"@"+timeconsume.String())
	go func() {
		cophd.rp.Subscribe(pubsub1, 0, func(msg string) int {
			if msg == hi {
				cophd.rp.Notify(channelname, cophdid+"@"+mtd+"@"+timeconsume.String())
			}
			return 0
		})
	}()
	go func() {
		cophd.rp.Subscribe(pubsub2, 0, func(msg string) int {
			copdto := cooperationDto{}
			err := json.Unmarshal([]byte(msg), &copdto)
			if err == nil {
				if len(receivebuffer) == cophd.buffersize {
					cophd.rp.Notify(copdto.CallerChannel, full)
					return 0
				}
				receivebuffer <- copdto
			}
			return 0
		})
	}()
	go func() {
		for {
			copdto := <-sendbuffer
			b, err := json.Marshal(copdto)
			if err == nil {
				cophd.rp.Notify(copdto.CallerChannel, string(b))
			}
		}
	}()
	for i := 0; i < cophd.concurrency; i++ {
		go func() {
			for {
				copdto := <-receivebuffer
				t := Utils.NewTimer()
				res := handler(copdto.Data)
				if res == nil {
					continue
				}
				d, ok := res.([]byte)
				if ok {
					copdto.Data = d
					sendbuffer <- copdto
				} else {
					d, err := json.Marshal(res)
					if err == nil {
						copdto.Data = d
						sendbuffer <- copdto
					}
				}
				tc := t.Stamp()
				if timeconsume == 0 {
					timeconsume = tc
				} else {
					timeconsume += (tc - timeconsume) / time.Duration(cophd.concurrency)
				}
			}
		}()
	}
}
