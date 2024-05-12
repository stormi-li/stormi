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
	name         string
}

func (cop *CooperationProxy) NewHandler() *CooperationHandler {
	cophd := CooperationHandler{}
	cophd.coprotocolId = cop.uuid
	cophd.concurrency = 10
	cophd.buffersize = 1000
	cophd.rp = cop.cp.rp
	cophd.name = cop.cooperationName
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
	TimeRemaining time.Duration
}

func (cophd *CooperationHandler) Handle(method int, handler func(data []byte) any) {
	StormiFmtPrintln(green, cophd.rp.addrs[0], "协作处理程序启动, 协议名:", cophd.name, "请求码:", method)
	receivebuffer := make(chan cooperationDto, cophd.buffersize)
	channelname := cophd.coprotocolId
	cophdid := uuid.NewString()
	pubsub := cophd.rp.GetPubSub(cophdid)
	var timeconsume time.Duration
	mtd := strconv.Itoa(method)
	go func() {
		for {
			cophd.rp.Notify(channelname, cophdid+"@"+mtd)
			time.Sleep(1 * time.Second)
		}
	}()
	go func() {
		cophd.rp.Subscribe(pubsub, 0, func(msg string) int {
			copdto := cooperationDto{}
			err := json.Unmarshal([]byte(msg), &copdto)
			if err == nil {
				if len(receivebuffer) == cophd.buffersize {
					copdto.Data = nil
					j, _ := json.Marshal(copdto)
					cophd.rp.Notify(copdto.CallerChannel, string(j))
					return 0
				}

				if copdto.TimeRemaining < timeconsume*time.Duration((len(receivebuffer)/cophd.concurrency)+1)+100*time.Millisecond {
					copdto.Data = nil
					j, _ := json.Marshal(copdto)
					cophd.rp.Notify(copdto.CallerChannel, string(j))
					return 0
				}
				receivebuffer <- copdto
			}
			return 0
		})
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
				} else {
					d, err := json.Marshal(res)
					if err == nil {
						copdto.Data = d
						b, err := json.Marshal(copdto)
						if err == nil {
							cophd.rp.Notify(copdto.CallerChannel, string(b))
						}
					}
				}
				tc := t.Stamp()
				if timeconsume == 0 {
					timeconsume = tc
				} else {
					timeconsume = tc
				}
			}
		}()
	}
}
