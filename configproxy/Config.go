package configproxy

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/stormi-li/stormi/dirandfileopt"
	"gopkg.in/yaml.v2"
)

type StormiConfig struct {
	Stormi struct {
		Ip    string `yaml:"ip"`
		Redis struct {
			Node  string `yaml:"node"`
			Nodes string `yaml:"nodes"`
		}
	}
}

const (
	reset   = "\x1b[0m"
	red     = "\x1b[31m"
	green   = "\x1b[32m"
	yellow  = "\x1b[33m"
	blue    = "\x1b[34m"
	magenta = "\x1b[35m"
	cyan    = "\x1b[36m"
)

var Config = StormiConfig{}
var ConfigMap map[string][][]string

// var modDir = serverSetDir()
var currentDir string

func ConfigProxy() {
	currentDir, _ = os.Getwd()
	yamlFile, _ := os.ReadFile(currentDir + "/app.yaml")
	yaml.Unmarshal(yamlFile, &Config)
	ConfigInit()
}

func FormatTime() string {
	return "[" + time.Now().Truncate(time.Second).Format("2006-01-02 15:04:05") + "]:"
}

func ConfigInit() {
	json, _ := json.MarshalIndent(Config, " ", " ")
	fmt.Println(cyan+FormatTime(), "加载配置:\n", string(json), reset)
	redisInit()
	configSync()
	autoSyncConfig()
}

var prefix = "stormi:config:"

func configSync() {
	fmt.Println(magenta+FormatTime(), "开始同步配置", reset)
	yamlFile, _ := os.ReadFile(currentDir + "/app.yaml")
	yaml.Unmarshal(yamlFile, &Config)
	uploadRedisClusterNode()
}

func uploadRedisClusterNode() {
	ctx := context.Background()
	rdsClusterClient.SAdd(ctx, prefix+"register", "redis-nodes")
	result, err := rdsClusterClient.ClusterNodes(ctx).Result()
	if err != nil {
		fmt.Println("获取 Redis 集群节点信息失败：", err)
		return
	} else {
		rdsClusterClient.Del(ctx, prefix+"addr:redis-nodes")
	}

	for _, line := range strings.Split(result, "\n") {
		fields := strings.Fields(line)
		if len(fields) > 1 && fields[len(fields)-1] != "disconnected" {
			addressParts := strings.Split(fields[1], "@")
			ipPort := strings.Split(addressParts[0], ":")
			if len(ipPort) != 2 {
				continue
			}
			node := ipPort[0] + ":" + ipPort[1]
			exist, _ := rdsClusterClient.SIsMember(ctx, prefix+"ignore", node).Result()
			if !exist {
				rdsClusterClient.HSet(ctx, prefix+"addr:redis-nodes", node, "redis集群节点")
			}
		}
	}
	downloadConfig()
}

func downloadConfig() {
	ConfigMap = make(map[string][][]string)
	var content = []string{}
	ctx := context.Background()
	res, _ := rdsClusterClient.SMembers(ctx, prefix+"register").Result()
	if len(res) > 0 {
		for _, name := range res {
			hmap, _ := rdsClusterClient.HGetAll(ctx, prefix+"addr:"+name).Result()
			name = strings.ReplaceAll(strings.ReplaceAll(name, "<", "("), ">", ")")
			if len(hmap) > 0 {
				for addr, desc := range hmap {
					if addr == "" || name == "" {
						continue
					}
					exist, _ := rdsClusterClient.SIsMember(ctx, prefix+"ignore", addr).Result()
					if !exist {
						ConfigMap[name] = append(ConfigMap[name], []string{addr, desc})
						content = append(content, name+"@"+addr+"@"+strings.ReplaceAll(strings.ReplaceAll(desc, "<", "("), ">", ")"))
					}
				}
			}
		}
	}
	dirandfileopt.WriteToConfigFile(content)
	var nodes []string
	for _, node := range ConfigMap["redis-nodes"] {
		nodes = append(nodes, node[0])
	}
	dirandfileopt.AppendToYaml(currentDir+"/app.yaml", nodes)
	fmt.Println(magenta+FormatTime(), "配置同步完成", reset)
	reInit()
}

func reInit() {
	fmt.Println(blue+FormatTime(), "开始连接最新redis集群和nsqd集群节点", reset)
	redisInit()
	fmt.Println(blue+FormatTime(), "已连接最新redis集群和nsqd集群节点", reset)
	if len(configHandlers) > 0 {
		fmt.Println(green+FormatTime(), "执行自定义配置处理", reset)
		for name, handler := range configHandlers {
			handler(ConfigMap[name])
		}
		fmt.Println(green+FormatTime(), "自定义配置处理执行完毕", reset)
	}
}

func (StormiConfig) RegisterConfig(name, addr, desc string) {
	ctx := context.Background()
	rdsClusterClient.SAdd(ctx, prefix+"register", name)
	rdsClusterClient.SRem(ctx, prefix+"ignore", addr)
	rdsClusterClient.HSet(ctx, prefix+"addr:"+name, addr, desc)
}

func (StormiConfig) IgnoreConfig(name, addr string) {
	ctx := context.Background()
	rdsClusterClient.SAdd(ctx, prefix+"ignore", addr)
	rdsClusterClient.HDel(ctx, prefix+"addr:"+name, addr)
}

func (StormiConfig) NotifySyncConfig() {
	notify("SyncConfig", "sync")
}

func autoSyncConfig() {
	go func() {
		for {
			cycleWait("SyncConfig", time.Hour, func(msg string) {
				if msg == "" {
					fmt.Println(yellow+FormatTime(), "定时同步配置任务触发", reset)
					configSync()
				} else {
					fmt.Println(yellow+FormatTime(), "接收到同步配置通知", reset)
					configSync()
				}
			})
		}
	}()
}

var configHandlers = make(map[string]func(nodes [][]string))

func (StormiConfig) AddConfigHandler(name string, handler func(nodes [][]string)) {
	configHandlers[name] = handler
}
