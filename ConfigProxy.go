package stormi

import (
	"context"
	"encoding/json"
	"time"

	"github.com/go-redis/redis/v8"
	"github.com/google/uuid"
)

type ConfigHandler struct {
	Name    string
	Handler func(config map[string]*Config)
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

func (c Config) GetKey() string {
	return c.Addr + "@" + c.UUID
}

func (c Config) GetRedisKey() string {
	return configAddrPrefix + c.Addr + "@" + c.UUID
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

type ConfigProxy struct {
	rp             *RedisProxy
	rdsAddr        string
	configHandlers []ConfigHandler
	ConfigSet      map[string]map[string]*Config
	count          int
}

func NewConfigProxy(addr any) *ConfigProxy {
	cp := &ConfigProxy{}
	cp.rp = NewRedisProxy(addr)
	cp.configHandlers = []ConfigHandler{}
	cp.ConfigSet = map[string]map[string]*Config{}
	cp.rdsAddr = cp.rp.addrs[0]
	cp.count = 0
	cp.autoSyncConfig()
	return cp
}

func (cp ConfigProxy) NewConfig() Config {
	c := Config{}
	c.UUID = uuid.NewString()
	return c
}

func (cp ConfigProxy) Info() {
	if !cp.rp.isConnected {
		StormiFmtPrintln(magenta, cp.rdsAddr, "当前未连接到任何配置集redis节点")
	}
	if cp.rp.isCluster {
		StormiFmtPrintln(yellow, cp.rdsAddr, "当前连接到redis集群节点:", cp.rdsAddr, "配置集正常工作")
	} else {
		StormiFmtPrintln(green, cp.rdsAddr, "当前连接到redis单例节点:", cp.rdsAddr, "配置集正常工作, 建议连接到redis集群")
	}
}

func (cp ConfigProxy) IsConnected() bool {
	return cp.rp.isConnected
}

func (cp ConfigProxy) IsCluster() bool {
	return cp.rp.isCluster
}

func (cp ConfigProxy) GetRdsClient() *redis.Client {
	return cp.rp.rdsClient
}

func (cp ConfigProxy) GetRdsClusterClient() *redis.ClusterClient {
	return cp.rp.rdsClusterClient
}

func (cp ConfigProxy) GetRdsAddr() string {
	return cp.rdsAddr
}

func (cp ConfigProxy) PrintConfigSet() {
	if len(cp.ConfigSet) == 0 {
		StormiFmtPrintln(magenta, cp.rdsAddr, "当前配置集为空")
	}
	StormiFmtPrintln(green, cp.rdsAddr, "当前配置集: ")
	for _, cmap := range cp.ConfigSet {
		for _, c := range cmap {
			StormiFmtPrintln(white, cp.rdsAddr, c.ToJsonStr())
		}
	}
}

func (cp ConfigProxy) ConfigsToAddrs(cs map[string]Config) []string {
	var addrs []string
	for _, c := range cs {
		addrs = append(addrs, c.Addr)
	}
	return addrs
}

const configRegisterPrefix = "stormi:config:register"
const configAddrPrefix = "stormi:config:addr:"

func (cp ConfigProxy) Register(c Config) {
	if cp.IsExist(c) {
		StormiFmtPrintln(magenta, cp.rdsAddr, "配置已存在:", c.ToString())
		return
	}
	cp.uploadConfig(c, "注册")
}

func (cp ConfigProxy) UpdateConfig(c Config) {
	cp.uploadConfig(c, "修改")
}

func (cp ConfigProxy) RefreshConfigs(cs []Config) {
	for _, c := range cs {
		cp.uploadConfig(c, "")
	}
}

func (cp ConfigProxy) uploadConfig(c Config, s string) {
	if c.Name == "" || c.Addr == "" || c.UUID == "" {
		StormiFmtPrintln(magenta, cp.rdsAddr, "配置信息不完全", c.ToConfigString())
	}
	var ok int64
	if cp.rp.isCluster {
		cp.rp.rdsClusterClient.SAdd(context.Background(), configRegisterPrefix, c.Name)
		ok, _ = cp.rp.rdsClusterClient.HSet(context.Background(), configAddrPrefix+c.Name, c.Addr+"@"+c.UUID, c.ToString()).Result()
	} else {
		cp.rp.rdsClient.SAdd(context.Background(), configRegisterPrefix, c.Name)
		ok, _ = cp.rp.rdsClient.HSet(context.Background(), configAddrPrefix+c.Name, c.Addr+"@"+c.UUID, c.ToString()).Result()
	}
	if ok != 0 {
		if s == "" {
			return
		}
		StormiFmtPrintln(yellow, cp.rdsAddr, "配置"+s+"成功, 新增配置:", c.ToConfigString())
	}
}

func (cp ConfigProxy) PullConfig(name string) map[string]Config {
	cmap := map[string]Config{}
	var hmap map[string]string
	if cp.rp.isCluster {
		hmap, _ = cp.rp.rdsClusterClient.HGetAll(context.Background(), configAddrPrefix+name).Result()
	} else {
		hmap, _ = cp.rp.rdsClient.HGetAll(context.Background(), configAddrPrefix+name).Result()
	}
	if len(hmap) > 0 {
		for n, cjson := range hmap {
			var c Config
			json.Unmarshal([]byte(cjson), &c)
			cmap[n] = c
		}
		return cmap
	}
	StormiFmtPrintln(magenta, cp.rdsAddr, "配置名不存在任何配置信息, name:", name, "建议在register里删除")
	return nil
}

func (cp ConfigProxy) PullAllConfig() map[string]map[string]Config {
	var names = []string{}
	var cmapmap = map[string]map[string]Config{}
	if cp.rp.isCluster {
		names, _ = cp.rp.rdsClusterClient.SMembers(context.Background(), configRegisterPrefix).Result()
	} else {
		names, _ = cp.rp.rdsClient.SMembers(context.Background(), configRegisterPrefix).Result()
	}
	if len(names) == 0 {
		StormiFmtPrintln(magenta, cp.rdsAddr, "redis配置集无任何配置信息")
		return nil
	}
	for _, name := range names {
		cmap := cp.PullConfig(name)
		cmapmap[name] = cmap
	}
	return cmapmap
}

func (cp ConfigProxy) IsExist(c Config) bool {
	var exist string
	if cp.rp.isCluster {
		exist, _ = cp.rp.rdsClusterClient.HGet(context.Background(), configAddrPrefix+c.Name, c.Addr+"@"+c.UUID).Result()
	} else {
		exist, _ = cp.rp.rdsClient.HGet(context.Background(), configAddrPrefix+c.Name, c.Addr+"@"+c.UUID).Result()
	}
	if exist != "" {
		return true
	}
	return false
}

var configchannel = "stormi-sync-config"

func (cp *ConfigProxy) SyncConfig() {
	if !cp.rp.isConnected {
		StormiFmtPrintln(red, cp.rdsAddr, "当前未连接到任何redis节点, 无法进行配置同步")
		return
	}
	StormiFmtPrintln(cyan, cp.rdsAddr, "开始第", cp.count+1, "次同步配置")
	cmapmap := cp.PullAllConfig()
	for name, cmap := range cmapmap {
		for key, c := range cmap {
			if cp.ConfigSet[name] == nil {
				cp.ConfigSet[name] = map[string]*Config{}
			}
			cp.ConfigSet[name][key] = &c
		}
	}
	for _, chandler := range cp.configHandlers {
		if cp.ConfigSet[chandler.Name] != nil {
			StormiFmtPrintln(green, cp.rdsAddr, "对配置", chandler.Name, "执行的自定义配置处理")
			chandler.Handler(cp.ConfigSet[chandler.Name])
			StormiFmtPrintln(green, cp.rdsAddr, "对配置", chandler.Name, "的自定义配置处理结束")
		}
	}
	StormiFmtPrintln(cyan, cp.rdsAddr, "第", cp.count+1, "次配置同步结束")
	cp.count++
}

func (cp ConfigProxy) autoSyncConfig() {
	if !cp.rp.isConnected {
		StormiFmtPrintln(red, cp.rdsAddr, "当前未连接到任何redis节点, 自动同步配置失败")
		return
	}
	StormiFmtPrintln(yellow, cp.rdsAddr, "初始配置同步开始")
	cp.SyncConfig()
	cp.PrintConfigSet()
	go func() {
		StormiFmtPrintln(yellow, cp.rdsAddr, "配置同步协程启动")
		cp.rp.CycleWait(configchannel, 1*time.Hour, func(msg string) {
			if msg == "" {
				StormiFmtPrintln(blue, cp.rdsAddr, "定时同步配置任务触发")
				cp.SyncConfig()
			} else {
				StormiFmtPrintln(blue, cp.rdsAddr, "接收到配置同步通知, 通知内容为:", msg)
				if notifyhandler == nil {
					cp.SyncConfig()
				} else {
					notifyhandler(cp, msg)
				}
			}
		})
	}()
}

var notifyhandler func(configProxy ConfigProxy, msg string)

func (ConfigProxy) SetConfigSyncNotficationHandler(handler func(configProxy ConfigProxy, msg string)) {
	notifyhandler = handler
}

func (cp ConfigProxy) NotifySync(desc string) {
	cp.rp.Notify(configchannel, desc)
}

func (cp *ConfigProxy) AddConfigHandler(name string, handler func(cmap map[string]*Config)) {
	ch := ConfigHandler{}
	ch.Name = name
	ch.Handler = handler
	cp.configHandlers = append(cp.configHandlers, ch)
}

func (cp *ConfigProxy) RemoveConfig(c Config) {
	var ok int64
	if cp.rp.isCluster {
		ok, _ = cp.rp.rdsClusterClient.HDel(context.Background(), configAddrPrefix+c.Name, c.Addr+"@"+c.UUID).Result()
	} else {
		ok, _ = cp.rp.rdsClient.HDel(context.Background(), configAddrPrefix+c.Name, c.Addr+"@"+c.UUID).Result()
	}
	if ok != 0 {
		StormiFmtPrintln(blue, cp.rdsAddr, "删除配置成功:", c.ToJsonStr())
	} else {
		StormiFmtPrintln(blue, cp.rdsAddr, "删除配置失败, 可能是该配置不存在于控制集:", c.ToJsonStr())
	}
}

func (cp *ConfigProxy) RemoveConfigs(cs []Config) {
	for _, c := range cs {
		cp.RemoveConfig(c)
	}
}
