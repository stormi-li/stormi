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

type ConfigProxy struct {
	rp             *RedisProxy
	rdsAddr        string
	configHandlers []ConfigHandler
	ConfigSet      map[string]map[string]*Config
	count          int
}

func (cp ConfigProxy) RedisProxy() *RedisProxy {
	return cp.rp
}

func NewConfigProxy(rp *RedisProxy) *ConfigProxy {
	cp := ConfigProxy{}
	cp.rp = rp
	cp.configHandlers = []ConfigHandler{}
	cp.ConfigSet = map[string]map[string]*Config{}
	cp.rdsAddr = cp.rp.addrs[0]
	cp.autoSync()
	return &cp
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
	cp.upload(c, "注册")
}

func (cp ConfigProxy) Update(c Config) {
	if cp.IsExist(c) {
		cp.upload(c, "修改")
	}
}

func (cp ConfigProxy) Refreshs(cset any) {
	cs, ok := cset.([]Config)
	if ok {
		for _, c := range cs {
			cp.upload(c, "")
		}
		return
	}
	cmap, ok := cset.(map[string]Config)
	if ok {
		for _, c := range cmap {
			cp.upload(c, "")
		}
		return
	}
	StormiFmtPrintln(magenta, cp.rdsAddr, "删除批量刷新配置失败, 只接受[]Config类型和map[string]Config类型")
}

func (cp ConfigProxy) upload(c Config, s string) {
	var ok int64
	if c.Name != "" {
		if cp.rp.isCluster {
			cp.rp.rdsClusterClient.SAdd(context.Background(), configRegisterPrefix, c.Name)
		} else if cp.rp.isConnected {
			cp.rp.rdsClient.SAdd(context.Background(), configRegisterPrefix, c.Name)
		}
	}
	if c.Addr == "" || c.UUID == "" {
		return
	}
	if cp.rp.isCluster {
		ok, _ = cp.rp.rdsClusterClient.HSet(context.Background(), configAddrPrefix+c.Name, c.Addr+"@"+c.UUID, c.ToString()).Result()
	} else if cp.rp.isConnected {
		ok, _ = cp.rp.rdsClient.HSet(context.Background(), configAddrPrefix+c.Name, c.Addr+"@"+c.UUID, c.ToString()).Result()
	}
	if ok != 0 {
		if s == "" {
			StormiFmtPrintln(yellow, cp.rdsAddr, "配置刷新成功, 刷新配置:", c.ToJsonStr())
			return
		}
		StormiFmtPrintln(yellow, cp.rdsAddr, "配置"+s+"成功, 新增配置:", c.ToJsonStr())
	}
}

func (cp ConfigProxy) Pull(name string) map[string]Config {
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
	return nil
}

func (cp ConfigProxy) PullAll() map[string]map[string]Config {
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
		cmap := cp.Pull(name)
		cmapmap[name] = cmap
	}
	return cmapmap
}

func (cp ConfigProxy) IsExist(c Config) bool {
	var exist string
	if cp.rp.isCluster {
		exist, _ = cp.rp.rdsClusterClient.HGet(context.Background(), configAddrPrefix+c.Name, c.Addr+"@"+c.UUID).Result()
	} else if cp.rp.isConnected {
		exist, _ = cp.rp.rdsClient.HGet(context.Background(), configAddrPrefix+c.Name, c.Addr+"@"+c.UUID).Result()
	}
	if exist != "" {
		return true
	}
	return false
}

func (cp ConfigProxy) IsRegistered(name string) bool {
	var exist bool
	if cp.rp.isCluster {
		exist, _ = cp.rp.rdsClusterClient.SIsMember(context.Background(), configRegisterPrefix, name).Result()
	} else if cp.rp.isConnected {
		exist, _ = cp.rp.rdsClient.SIsMember(context.Background(), configRegisterPrefix, name).Result()
	}
	if exist {
		return true
	}
	return false
}

var configchannel = "stormi-sync-config"

func (cp *ConfigProxy) Sync() {
	if !cp.rp.isConnected {
		StormiFmtPrintln(red, cp.rdsAddr, "当前未连接到任何redis节点, 无法进行配置同步")
		return
	}
	cp.count = cp.count + 1
	StormiFmtPrintln(cyan, cp.rdsAddr, "开始第", cp.count, "次同步配置")
	cmapmap := cp.PullAll()
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
	StormiFmtPrintln(cyan, cp.rdsAddr, "第", cp.count, "次配置同步结束")
}

func (cp *ConfigProxy) autoSync() {
	if !cp.rp.isConnected {
		StormiFmtPrintln(red, cp.rdsAddr, "当前未连接到任何redis节点, 自动同步配置失败")
		return
	}
	StormiFmtPrintln(yellow, cp.rdsAddr, "初始配置同步开始")
	cp.Sync()
	cp.PrintConfigSet()
	StormiFmtPrintln(yellow, cp.rdsAddr, "配置同步协程启动")
	go func() {
		cp.rp.CycleWait(configchannel, 1*time.Hour, false, func(msg *string) error {
			if msg == nil {
				StormiFmtPrintln(blue, cp.rdsAddr, "定时同步配置任务触发")
				cp.Sync()
			} else {
				StormiFmtPrintln(blue, cp.rdsAddr, "接收到配置同步通知, 通知内容为:", *msg)
				if notifyhandlers == nil {
					cp.Sync()
				} else {
					for _, notifyhandler := range notifyhandlers {
						notifyhandler(*cp, *msg)
					}
				}
			}
			return nil
		})
	}()
}

var notifyhandlers []func(configProxy ConfigProxy, msg string)

func (ConfigProxy) AddConfigSyncNotificationHandler(handler func(configProxy ConfigProxy, msg string)) {
	notifyhandlers = append(notifyhandlers, handler)
}

func (cp ConfigProxy) NotifySync(desc string) {
	cp.rp.Notify(configchannel, desc)
}

func (cp *ConfigProxy) AddConfigHandler(name string, handler func(cmap map[string]*Config)) {
	ch := ConfigHandler{}
	ch.Name = name
	ch.Handler = handler
	cp.configHandlers = append(cp.configHandlers, ch)
	cp.Sync()
}

func (cp *ConfigProxy) Remove(c Config) {
	var ok int64
	if cp.rp.isCluster {
		ok, _ = cp.rp.rdsClusterClient.HDel(context.Background(), configAddrPrefix+c.Name, c.Addr+"@"+c.UUID).Result()
	} else {
		ok, _ = cp.rp.rdsClient.HDel(context.Background(), configAddrPrefix+c.Name, c.Addr+"@"+c.UUID).Result()
	}
	if ok != 0 {
		StormiFmtPrintln(blue, cp.rdsAddr, "删除配置成功:", c.ToJsonStr())
	} else {
		StormiFmtPrintln(blue, cp.rdsAddr, "删除配置失败, 可能是该配置不存在于配置集:", c.ToJsonStr())
	}
}

func (cp *ConfigProxy) RemoveRegister(name string) {
	var ok int64
	if cp.rp.isCluster {
		ok, _ = cp.rp.rdsClusterClient.SRem(context.Background(), configRegisterPrefix, name).Result()
	} else {
		ok, _ = cp.rp.rdsClient.HDel(context.Background(), configRegisterPrefix, name).Result()
	}
	if ok != 0 {
		StormiFmtPrintln(blue, cp.rdsAddr, "删除配置名成功:", name)
	} else {
		StormiFmtPrintln(blue, cp.rdsAddr, "删除配置名失败, 可能是该配置名不存在于注册集:", name)
	}
}

func (cp *ConfigProxy) Removes(cset any) {
	cs, ok := cset.([]Config)
	if ok {
		for _, c := range cs {
			cp.Remove(c)
		}
		return
	}
	cmap, ok := cset.(map[string]Config)
	if ok {
		for _, c := range cmap {
			cp.Remove(c)
		}
		return
	}
	StormiFmtPrintln(magenta, cp.rdsAddr, "删除批量删除配置失败, 只能删除[]Config类型和map[string]Config类型")
}

func (cp *ConfigProxy) RegisterRedisStandalone(nodeId int) {
	c := cp.NewConfig()
	c.Name = "redis-single"
	c.Addr = cp.rp.addrs[0]
	c.NodeId = nodeId
	cp.Register(c)
}

func (cp *ConfigProxy) RegisterRedisClusterNode() {
	for _, addr := range cp.rp.addrs {
		c := cp.NewConfig()
		c.Name = "redis-cluster"
		c.Addr = addr
		cp.Register(c)
	}
}

func (cp *ConfigProxy) ConfigPersistence() {
	var err error
	if cp.rp.isCluster {
		err = cp.rp.rdsClusterClient.BgSave(context.Background()).Err()
	} else if cp.rp.isConnected {
		err = cp.rp.rdsClient.BgSave(context.Background()).Err()
	} else {
		StormiFmtPrintln(yellow, cp.rdsAddr, "未连接到redis, 持久化配置失败")
	}
	if err == nil {
		StormiFmtPrintln(yellow, cp.rdsAddr, "redis持久化配置成功")
	}
}
