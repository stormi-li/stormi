package stormi

import (
	"math/rand"
	"net"
	"strconv"
	"time"
)

type ServerProxy struct {
	name       string
	serverAddr string
	weight     int
	configs    []Config
	rdsAddr    string
	cp         *ConfigProxy
	sdreg      chan struct{}
	sddis      chan struct{}
}

func NewServerProxy(addr any) *ServerProxy {
	sp := ServerProxy{}
	sp.cp = NewConfigProxy(addr)
	sp.rdsAddr = sp.cp.rdsAddr
	sp.sdreg = make(chan struct{})
	return &sp
}

func (sp *ServerProxy) Register(name string, addr string, weight int, t time.Duration) {
	StormiFmtPrintln(green, sp.rdsAddr, "注册服务启动, 服务名:", name, "地址:", addr, "权重:", weight, "心跳间隔:", t)
	for i := 0; i < weight; i++ {
		c := sp.cp.NewConfig()
		c.Name = name
		c.Addr = addr
		sp.cp.Register(c)
		sp.configs = append(sp.configs, c)
	}
	sp.cp.NotifySync("发布服务, 服务名:" + name + "地址:" + addr + "权重:" + strconv.Itoa(weight) + "心跳间隔:" + t.String())
	rp := sp.cp.rp
	var msgc = make(chan string, 1)
	var sdpub = make(chan struct{})
	var sdcyc = make(chan struct{})
	go func() {
		rp.Publish(name+addr, msgc, sdpub)
	}()
	go func() {
		cycleTask(t, func(msg string) {
			msgc <- t.String()
			sp.cp.RefreshConfigs(sp.configs)
		}, sdcyc)
	}()
	go func() {
		StormiFmtPrintln(green, sp.rdsAddr, "开始监听客户端同步请求")
		sp.cp.rp.Subscribe(sp.cp.rp.GetSubChannel(name+addr+"server"), 0, func(msg string) int {
			StormiFmtPrintln(yellow, sp.rdsAddr, "接收到客户端同步请求:", msg)
			msgc <- t.String()
			return 0
		})
	}()
	go func() {
		<-sp.sdreg
		sdpub <- struct{}{}
		sdcyc <- struct{}{}
		sp.cp.RemoveConfigs(sp.configs)
		StormiFmtPrintln(green, sp.rdsAddr, "注册服务关闭, 服务名:", name, "地址:", addr, "权重:", weight, "心跳间隔:", t)
	}()
}

func cycleTaskDelay(t time.Duration, handler func(msg string), sd chan struct{}) {
	timer := time.NewTicker(t)
	for {
		select {
		case <-timer.C:
			timer = time.NewTicker(t)
			handler("")
		case <-sd:
			return
		}
	}
}
func cycleTask(t time.Duration, handler func(msg string), sd chan struct{}) {
	handler("")
	cycleTaskDelay(t, handler, sd)
}

func (sp *ServerProxy) Shutdown() {
	sp.sdreg <- struct{}{}
}

func (sp *ServerProxy) Discover(name string, t time.Duration, handler func(addr string) error) {
	c := sp.discover(name)
	if c == nil {
		StormiFmtPrintln(magenta, sp.rdsAddr, "当前配置集未发现", name, "服务, 尝试从redis配置集重新拉取配置")
		sp.cp.PullAllConfig()
		c = sp.discover(name)
		if c == nil {
			StormiFmtPrintln(magenta, sp.rdsAddr, "redis配置集未发现", name, "服务, 发现服务关闭")
			return
		}
	}
	pubName := c.Name + c.Addr
	ch := sp.cp.rp.GetSubChannel(pubName)
	sp.cp.rp.Notify(pubName+"server", getIp())
	var heart string
	res := sp.cp.rp.Subscribe(ch, t, func(msg string) int {
		if msg != "" {
			heart = msg
			return 1
		}
		return 2
	})
	if res != 1 {
		sp.cp.ConfigSet[name][c.GetKey()].Ignore = true
		sp.cp.UpdateConfig(*sp.cp.ConfigSet[name][c.GetKey()])
		sp.Discover(name, t, handler)
		return
	}

	StormiFmtPrintln(yellow, sp.rdsAddr, "发现服务: ", c.ToJsonStr(), "开始监控其心跳")
	go func() {
		duration, err := time.ParseDuration(heart)
		if err != nil {
			sp.Discover(name, t, handler)
			return
		}
		if handler(c.Addr) != nil {
			StormiFmtPrintln(magenta, sp.rdsAddr, "服务处理错误: ", err.Error(), "重新拉取新服务")
			sp.Discover(name, t, handler)
			return
		}
		for {
			h := sp.cp.rp.Wait(c.Name+c.Addr, duration*2)
			if h == "" {
				StormiFmtPrintln(magenta, sp.rdsAddr, "服务断联: ", c.ToJsonStr(), "尝试重新拉取新服务")
				break
			}
		}
		sp.Discover(name, t, handler)
	}()
}

func (sp *ServerProxy) discover(name string) *Config {
	cmap := sp.cp.ConfigSet[name]
	if len(cmap) == 0 {
		return nil
	}
	for {
		if len(cmap) == 0 {
			return nil
		}
		k, c := randMapSC(cmap)
		if c.Ignore {
			delete(cmap, k)
			continue
		}
		return &c
	}
}

func randMapSC(m map[string]*Config) (string, Config) {
	keys := make([]string, 0, len(m))
	for key := range m {
		keys = append(keys, key)
	}

	randomIndex := rand.Intn(len(keys))
	randomKey := keys[randomIndex]

	randomValue := m[randomKey]

	return randomKey, *randomValue
}

func getIp() string {
	addrs, _ := net.InterfaceAddrs()
	return addrs[1].String()
}
