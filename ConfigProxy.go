package stormi

import (
	"context"
	"encoding/json"
	"time"

	"github.com/google/uuid"
)

var configHandlers = []ConfigHandler{}
var ConfigSet = map[string]map[string]Config{}
var count int

type ConfigHandler struct {
	Name    string
	Handler func(config map[string]Config)
}

type Config struct {
	Name   string
	Addr   string
	UUID   string
	NodeId int
	Desc   string
	Info   map[string]string
	Ignore bool
}

func (c Config) ToString() string {
	bs, _ := json.MarshalIndent(c, " ", "  ")
	return string(bs)
}

func (c Config) ToJson() []byte {
	bs, _ := json.Marshal(c)
	return bs
}

func (c Config) ToJsonStr() string {
	bs, _ := json.Marshal(c)
	return string(bs)
}

func NewConfig() Config {
	var c = Config{}
	c.UUID = uuid.NewString()
	return c
}

type configOpt struct{}

var configProxy *configOpt

var addr string

func NewConfigProxy(addr string) *configOpt {
	addr = addr
	if configProxy != nil {
		return configProxy
	}
	redisInit(addr)
	if len(ConfigSet) > 0 {
		cs := ConfigSet["redis-cluster"]
		if len(cs) > 0 {
			addrs := configProxy.ConfigsToAddrs(cs)
			redisInit(addrs)
		}
	}
	configProxy = &configOpt{}
	autoSyncConfig()
	return configProxy
}

func (configOpt) Info() {
	if isConnected {
		StormiFmtPrintln(magenta, "当前未连接到任何配置集redis节点")
	}
	if isCluster {
		StormiFmtPrintln(yellow, "当前连接到redis集群节点:", addr, "配置集正常工作")
	}
	if isCluster {
		StormiFmtPrintln(green, "当前连接到redis单例节点:", addr, "配置集正常工作, 建议连接到redis集群")
	}
}

func (configOpt) PrintConfigSet() {
	if len(ConfigSet) == 0 {
		StormiFmtPrintln(magenta, "当前配置集为空")
	}
	StormiFmtPrintln(green, "配置集: ")
	for _, cmap := range ConfigSet {
		for _, c := range cmap {
			StormiFmtPrintln(green, c.ToJsonStr())
		}
	}
}

func (configOpt) ConfigsToAddrs(cs map[string]Config) []string {
	var addrs []string
	for _, c := range cs {
		addrs = append(addrs, c.Addr)
	}
	return addrs
}

const configPrefix = "stormi:config:"
const configRegisterPrefix = "stormi:config:register"
const configAddrPrefix = "stormi:config:addr:"

func (configOpt) RegisterConfig(c Config) {
	if configProxy.IsExist(c) {
		StormiFmtPrintln(magenta, "配置已存在:", c.ToString())
		return
	}
	configProxy.uploadConfig(c, "注册")
}

func (configOpt) UpdateConfig(c Config) {
	configProxy.uploadConfig(c, "修改")
}

func (configOpt) uploadConfig(c Config, s string) {
	if c.Name == "" || c.Addr == "" || c.UUID == "" {
		StormiFmtPrintln(magenta, "配置信息不完全", c.ToConfigString())
	}
	if isCluster {
		ctx := context.Background()
		rdsClusterClient.SAdd(ctx, configRegisterPrefix, c.Name)
		ok, _ := rdsClusterClient.HSet(ctx, configAddrPrefix+c.Name, c.Addr+"@"+c.UUID, c.ToString()).Result()
		if ok != 0 {
			StormiFmtPrintln(yellow, "配置"+s+"成功:", c.ToConfigString())
		}
	} else {
		ctx := context.Background()
		rdsClient.SAdd(ctx, configPrefix+"register", c.Name)
		rdsClient.SRem(ctx, configPrefix+"ignore", c.Addr+"@"+c.UUID)
		ok, _ := rdsClusterClient.HSet(ctx, configAddrPrefix+c.Name, c.Addr+"@"+c.UUID, c.ToString()).Result()
		if ok != 0 {
			StormiFmtPrintln(yellow, "配置"+s+"成功:", c.ToConfigString())
		}
	}
}

func (configOpt) PullConfig(name string) map[string]Config {
	cmap := map[string]Config{}
	var hmap map[string]string
	if isCluster {
		hmap, _ = rdsClusterClient.HGetAll(context.Background(), configAddrPrefix+name).Result()
	} else {
		hmap, _ = rdsClient.HGetAll(context.Background(), configAddrPrefix+name).Result()
	}
	if len(hmap) > 0 {
		for n, cjson := range hmap {
			var c Config
			json.Unmarshal([]byte(cjson), &c)
			cmap[n] = c
		}
		return cmap
	}
	StormiFmtPrintln(magenta, "配置名未在配置集, name:", name)
	return nil
}

func (configOpt) PullAllConfig() map[string]map[string]Config {
	var names = []string{}
	var cmapmap = map[string]map[string]Config{}
	if isCluster {
		names, _ = rdsClusterClient.SMembers(context.Background(), configRegisterPrefix).Result()
	} else {
		names, _ = rdsClient.SMembers(context.Background(), configRegisterPrefix).Result()
	}
	if len(names) == 0 {
		StormiFmtPrintln(magenta, "redis配置集无任何配置信息")
		return nil
	}
	for _, name := range names {
		cmap := configProxy.PullConfig(name)
		cmapmap[name] = cmap
	}
	return cmapmap
}

func (configOpt) IsExist(c Config) bool {
	var exist string
	if isCluster {
		exist, _ = rdsClusterClient.HGet(context.Background(), configAddrPrefix+c.Name, c.Addr+"@"+c.UUID).Result()
	} else {
		exist, _ = rdsClusterClient.HGet(context.Background(), configAddrPrefix+c.Name, c.Addr+"@"+c.UUID).Result()
	}
	if exist != "" {
		return true
	}
	return false
}

var configchannel = "stormi-sync-config"

func (configOpt) SyncConfig() {
	if !isConnected {
		StormiFmtPrintln(red, "当前未连接到任何redis节点, 无法进行配置同步")
		return
	}
	StormiFmtPrintln(cyan, "开始第", count+1, "次同步配置")
	cmapmap := configProxy.PullAllConfig()
	for name, cmap := range cmapmap {
		for key, c := range cmap {
			ConfigSet[name][key] = c
		}
	}
	for _, chandler := range configHandlers {
		if ConfigSet[chandler.Name] != nil {
			StormiFmtPrintln(cyan, "执行对配置名为", chandler.Name, "的自定义配置处理")
			chandler.Handler(ConfigSet[chandler.Name])
			StormiFmtPrintln(cyan, "对配置名为", chandler.Name, "的自定义配置处理结束")
		}
	}
	StormiFmtPrintln(cyan, "第", count+1, "次配置同步结束")
}

func autoSyncConfig() {
	if !isConnected {
		StormiFmtPrintln(red, "当前未连接到任何redis节点, 自动同步配置失败")
		return
	}
	StormiFmtPrintln(yellow, "初始配置同步开始")
	configProxy.SyncConfig()
	configProxy.PrintConfigSet()
	go func() {
		RedisProxy.CycleWait(configchannel, 1*time.Hour, func(msg string) {
			if msg == "" {
				StormiFmtPrintln(blue, "定时同步配置任务触发")
			} else {
				StormiFmtPrintln(blue, "接收到配置同步通知")
			}
			configProxy.SyncConfig()
		})
	}()
}

func (configOpt) NotifySync() {
	RedisProxy.Notify(configchannel, "sync")
}

func (configOpt) AddConfigHandler(name string, handler func(cmap map[string]Config)) {
	ch := ConfigHandler{}
	ch.Name = name
	ch.Handler = handler
	configHandlers = append(configHandlers, ch)
}
