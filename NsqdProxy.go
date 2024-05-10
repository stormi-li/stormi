package stormi

import (
	"math/rand"
	"strings"
	"time"

	"github.com/nsqio/go-nsq"
)

type NsqdProxy struct {
	cp             *ConfigProxy
	addrs          []string
	availableAddrs []string
	producers      []*nsq.Producer
	proReconnect   chan struct{}
	consReconnect  chan struct{}
	sdautocon      chan struct{}
	sdautocons     chan struct{}
	stoped         bool
}

func NewNsqdProxy(addr any) *NsqdProxy {
	cp, ok := addr.(*ConfigProxy)
	np := NsqdProxy{}
	if ok {
		np.cp = cp
	} else {
		np.cp = NewConfigProxy(addr)
	}
	np.proReconnect = make(chan struct{}, 1)
	np.consReconnect = make(chan struct{}, 1)
	np.sdautocon = make(chan struct{}, 1)
	np.sdautocons = make(chan struct{}, 1)
	np.refreshAddrs()
	np.cp.AddConfigSyncNotificationHandler(func(configProxy ConfigProxy, msg string) {
		parts := strings.Split(msg, "@")
		if len(parts) > 0 && parts[0] == "nsqd" {
			StormiFmtPrintln(green, np.cp.rdsAddr, "接收到关于nsqd节点的通知:", msg, "更新nsqd连接")
			configProxy.Sync()
			np.refreshAddrs()
			np.proReconnect <- struct{}{}
		}
	})
	np.autoConnect()
	np.autoConsume()
	return &np
}

func (np *NsqdProxy) Register(addr string) {
	c := np.cp.NewConfig()
	c.Name = "nsqd"
	c.Addr = addr
	c.Desc = "新增nsqd集群节点:" + addr
	np.cp.Register(c)
	np.cp.NotifySync("nsqd@" + c.Desc)
}

func (np *NsqdProxy) refreshAddrs() {
	np.addrs = []string{}
	nmap := np.cp.ConfigSet["nsqd"]
	if len(nmap) == 0 {
		StormiFmtPrintln(magenta, np.cp.rdsAddr, "当前配置集未发现nsqd节点, 尝试从redis重新拉取配置")
		np.cp.Sync()
		nmap = np.cp.ConfigSet["nsqd"]
		if len(nmap) == 0 {
			StormiFmtPrintln(magenta, np.cp.rdsAddr, "redis配置集未发现nsqd节点, 连接失败")
			return
		}
	}
	cs := []Config{}
	for _, c := range nmap {
		cs = append(cs, *c)
		if !c.Ignore {
			np.addrs = append(np.addrs, c.Addr)
		}
	}
	if len(np.addrs) == 0 {
		StormiFmtPrintln(magenta, np.cp.rdsAddr, "当前配置集未发现有效nsqd节点, 无效nsqd:")
		for _, c := range cs {
			StormiFmtPrintln(magenta, np.cp.rdsAddr, c.ToJsonStr())
		}
		return
	}
	StormiFmtPrintln(yellow, np.cp.rdsAddr, "更新nsqd节点结束, 当前配置集可供连接的nsqd节点:", np.addrs)
}

func (np *NsqdProxy) autoConnect() {
	go func() {
		cycleTaskWithTriger(10*time.Second, func() {
			if len(np.addrs) != 0 {
				changed := false
				for _, addr := range np.addrs {
					if np.isVailable(addr) {
						add := true
						for _, a := range np.availableAddrs {
							if a == addr {
								add = false
								break
							}
						}
						if add {
							c, _ := nsq.NewProducer(addr, nsq.NewConfig())
							c.SetLoggerLevel(nsq.LogLevelError)
							np.producers = append(np.producers, c)
							np.availableAddrs = append(np.availableAddrs, addr)
							StormiFmtPrintln(yellow, np.cp.rdsAddr, "已成功连接到的nsqd节点:", addr)
							np.consReconnect <- struct{}{}
							changed = true
						}
					}
				}
				if changed {
					if len(np.availableAddrs) == 0 {
						StormiFmtPrintln(magenta, np.cp.rdsAddr, "当前配置集nsqd节点均不可用")
					}
				}
			}
		}, np.proReconnect, np.sdautocon)
		if len(np.producers) != 0 {
			for _, p := range np.producers {
				p.Stop()
			}
		}
		np.sdautocons <- struct{}{}
	}()
}

func (np *NsqdProxy) isVailable(addr string) bool {
	c, _ := nsq.NewConsumer("connect-test", "test", nsq.NewConfig())
	defer c.Stop()
	c.SetLoggerLevel(nsq.LogLevelError)
	c.AddHandler(nsq.HandlerFunc(func(message *nsq.Message) error { return nil }))
	err := c.ConnectToNSQD(addr)
	if err != nil {
		return false
	} else {
		return true
	}
}

func (np *NsqdProxy) Publish(topic string, msg []byte) {
	for {
		if np.stoped {
			StormiFmtPrintln(magenta, np.cp.rdsAddr, "nsqd代理已关闭, 消息发送失败")
			return
		}
		if len(np.producers) == 0 {
			StormiFmtPrintln(magenta, np.cp.rdsAddr, "当前无可用nsqd节点, 等待一秒后重试")
			time.Sleep(1 * time.Second)
			continue
		}
		index := rand.Intn(len(np.producers))
		err := np.producers[index].Publish(topic, msg)
		if err == nil {
			break
		} else {
			np.producers[index].Stop()
			StormiFmtPrintln(magenta, np.cp.rdsAddr, "nsqd节点:"+np.availableAddrs[index], err, "已将其移出nsqd连接池, 等待自动重连")
			if index == len(np.producers)-1 {
				np.producers = np.producers[:index]
				np.availableAddrs = np.availableAddrs[:index]
			} else {
				np.producers = append(np.producers[:index], np.producers[index+1:]...)
				np.availableAddrs = append(np.availableAddrs[:index], np.availableAddrs[index+1:]...)
			}
			np.consReconnect <- struct{}{}
		}
	}
}

type consumeHanlder struct {
	topic   string
	channel string
	handler func(message *nsq.Message) error
}

var consumerHandlers []consumeHanlder

func (np *NsqdProxy) AddConsumeHandler(topic string, channel string, handler func(message *nsq.Message) error) {
	h := consumeHanlder{}
	h.topic = topic
	h.channel = channel
	h.handler = handler
	consumerHandlers = append(consumerHandlers, h)
	config := nsq.NewConfig()
	for _, node := range np.availableAddrs {
		c, err := nsq.NewConsumer(topic, channel, config)
		if err != nil {
			StormiFmtPrintln(magenta, np.cp.rdsAddr, "nsqd节点:"+node, err)
			continue
		}
		c.AddHandler(nsq.HandlerFunc(handler))
		err = c.ConnectToNSQD(node)
		if err != nil {
			StormiFmtPrintln(magenta, np.cp.rdsAddr, "nsqd节点:"+node, err)
		}
	}
}

func (np *NsqdProxy) autoConsume() {
	var consmap map[string][]*nsq.Consumer
	go func() {
		for {
			<-np.consReconnect
			if len(consumerHandlers) == 0 {
				continue
			}
			if len(consmap) != 0 {
				for _, cons := range consmap {
					for _, con := range cons {
						con.Stop()
					}
				}
			}
			consmap = make(map[string][]*nsq.Consumer)
			for _, addr := range np.availableAddrs {
				for _, ch := range consumerHandlers {
					c, err := nsq.NewConsumer(ch.topic, ch.channel, nsq.NewConfig())
					if err != nil {
						continue
					}
					c.SetLoggerLevel(nsq.LogLevelError)
					c.AddHandler(nsq.HandlerFunc(ch.handler))
					c.ConnectToNSQD(addr)
					consmap[addr] = append(consmap[addr], c)
				}
			}
		}
	}()
	go func() {
		<-np.sdautocons
		if len(consmap) != 0 {
			for _, cons := range consmap {
				for _, con := range cons {
					con.Stop()
				}
			}
		}
	}()
}

func (np *NsqdProxy) Stop() {
	np.sdautocon <- struct{}{}
	np.stoped = true
}

func cycleTaskWithTriger(t time.Duration, handler func(), triger, sd chan struct{}) {
	handler()
	cycleTaskDelayWithTriger(t, handler, triger, sd)
}

func cycleTaskDelayWithTriger(t time.Duration, handler func(), triger, sd chan struct{}) {
	timer := time.NewTicker(t)
	for {
		select {
		case <-timer.C:
			timer = time.NewTicker(t)
			handler()
		case <-triger:
			handler()
		case <-sd:
			return
		}
	}
}
