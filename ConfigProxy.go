package stormi

import (
	"context"
	"encoding/json"
	"strings"

	"github.com/google/uuid"
)

var redisnode string
var configHandlers map[string]ConfigHandler
var ConfigMap = map[string]map[string]Config{}

type ConfigHandler struct {
	Name    string
	Handler func(config Config)
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

func NewConfig() Config {
	var c = Config{}
	c.UUID = uuid.NewString()
	return c
}

type configOpt struct{}

var configProxy *configOpt

func NewConfigProxy(f string, addr string) *configOpt {
	if configProxy != nil {
		return configProxy
	}
	name := FileProxy.GetAvailableConfigFileName(f)
	filename = f + "/" + name + ".stormi"

	redisnode = addr
	ConfigMap = FileProxy.ReadConfigFile(filename)
	redisInit(addr)
	if len(ConfigMap) > 0 {
		cs := ConfigMap["redis-cluster"]
		if len(cs) > 0 {
			addrs := configProxy.ConfigsToAddrs(cs)
			redisInit(addrs)
		}
	}
	configProxy = &configOpt{}
	return configProxy
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

func (configOpt) PullConfig(name string) []Config {
	cs := []Config{}
	var hmap map[string]string
	if isCluster {
		hmap, _ = rdsClusterClient.HGetAll(context.Background(), configAddrPrefix+name).Result()
	} else {
		hmap, _ = rdsClient.HGetAll(context.Background(), configAddrPrefix+name).Result()
	}
	if len(hmap) > 0 {
		for n, cjson := range hmap {
			c := configProxy.stringToConfig(n, cjson)
			cs = append(cs, c)
		}
		return cs
	}
	StormiFmtPrintln(magenta, "配置名未在配置集, name:", name)
	return nil
}

func (configOpt) stringToConfig(s1, s2 string) Config {
	parts := strings.Split(s1, "@")
	c := Config{}
	json.Unmarshal([]byte(s2), &c)
	c.UUID = parts[1]
	return c
}

func (configOpt) jsonStringToConfig(s string) (string, Config) {
	parts := strings.Split(s, "@")
	if len(parts) != 4 {
		return "", Config{}
	}
	var config = Config{}
	json.Unmarshal([]byte(parts[3]), &config)
	return parts[0], config
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

func (configOpt) IsExist(c Config) bool {
	if isCluster {
		ctx := context.Background()
		rdsClusterClient.SAdd(ctx, configRegisterPrefix, c.Name)
		ok, _ := rdsClusterClient.HGet(ctx, configAddrPrefix+c.Name, c.Addr+"@"+c.UUID).Result()
		if ok != "" {
			return true
		}
	} else {
		ctx := context.Background()
		rdsClient.SAdd(ctx, configPrefix+"register", c.Name)
		rdsClient.SRem(ctx, configPrefix+"ignore", c.Addr+"@"+c.UUID)
		ok, _ := rdsClusterClient.HGet(ctx, configAddrPrefix+c.Name, c.Addr+"@"+c.UUID).Result()
		if ok != "" {
			return true
		}
	}
	return false
}

func (configOpt) SyncConfig() {

}

func (configOpt) NotifySync() {}

func (configOpt) AddConfigHandler() {}
