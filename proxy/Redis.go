package proxy

import (
	"context"
	"fmt"
	"math/rand"
	"os"
	"runtime"
	"strconv"
	"strings"
	"time"

	"github.com/go-redis/redis/v8"
	"github.com/google/uuid"
	"github.com/stormi-li/stormi/dirandfileopt"
	"github.com/stormi-li/stormi/formatprint"
)

var rdsClient *redis.Client
var rdsClusterClient *redis.ClusterClient

func redisInit() {
	if ConfigMap != nil {
		rdsClusterClient = redis.NewClusterClient(&redis.ClusterOptions{
			Addrs: ConfigMap["redis-nodes"][0],
		})
	} else {
		if Config.Stormi.Redis.Nodes != "" {
			redisNodestr := Config.Stormi.Redis.Nodes
			redisNodes := strings.Fields(redisNodestr)
			rdsClusterClient = redis.NewClusterClient(&redis.ClusterOptions{
				Addrs: redisNodes,
			})
		}
	}
	if Config.Stormi.Redis.Node != "" && rdsClient == nil {
		rdsClient = redis.NewClient(&redis.Options{
			Addr: Config.Stormi.Redis.Node,
		})
	}

}

func (redisOpt) RedisClient() *redis.Client {
	return rdsClient
}

func (redisOpt) RedisClusterClient() *redis.ClusterClient {
	return rdsClusterClient
}

func (redisOpt) RedisSingleNodeInfo() {
	opt := rdsClient.Options()

	fmt.Printf("Redis Single Node Address: %s\n", opt.Addr)
}

func (redisOpt) RedisClusterNodesInfo() {
	redisNodes, _ := rdsClusterClient.ClusterNodes(context.Background()).Result()
	fmt.Println("stormi redis redisNodes info:")
	fmt.Print(redisNodes)
}

type dLock struct {
	uuid     string
	lockName string
	stop     chan struct{}
}

type redisOpt struct{}

var RedisProxy redisOpt

func (redisOpt) NewLock(lockName string) *dLock {
	dLock := dLock{}
	dLock.lockName = lockName
	dLock.uuid = uuid.New().String()
	dLock.stop = make(chan struct{})
	return &dLock
}

func (l *dLock) Lock() {
	ctx := context.Background()
	for {
		var ok bool
		ok, _ = rdsClusterClient.SetNX(ctx, l.lockName, l.uuid, 3*time.Second).Result()

		if ok {
			go func() {
				ticker := time.NewTicker(1 * time.Second)
				defer ticker.Stop()
				for {
					select {
					case <-ticker.C:
						ok, _ = rdsClusterClient.SetNX(ctx, l.lockName, l.uuid, 3*time.Second).Result()
					case <-l.stop:
						return
					}
				}
			}()
			break
		} else {
			wait(l.lockName, 3*time.Second)
		}
	}
}

func (l *dLock) UnLock() {
	l.stop <- struct{}{}
	ctx := context.Background()
	var uuid string
	uuid, _ = rdsClusterClient.Get(ctx, l.lockName).Result()

	if uuid == l.uuid {
		rdsClient.Del(ctx, l.lockName)
		notify(l.lockName, "unlock")
	}
}

func notify(channel, msg string) {
	rdsClusterClient.Publish(context.Background(), channel, msg)
}

func wait(channel string, timeout time.Duration) string {
	pubsub := rdsClusterClient.Subscribe(context.Background(), channel)
	defer pubsub.Close()
	ch := pubsub.Channel()
	timer := time.NewTicker(timeout)
	defer timer.Stop()
	select {
	case <-timer.C:
		return ""
	case msg := <-ch:
		return msg.Payload
	}
}

func cycleWait(channel string, timeout time.Duration, handler func(msg string)) {
	pubsub := rdsClusterClient.Subscribe(context.Background(), channel)
	defer pubsub.Close()
	c := pubsub.Channel()
	timer := time.NewTicker(timeout)
	for {
		select {
		case <-timer.C:
			if timeout != 0 {
				handler("")
				timer = time.NewTicker(timeout)
			}
		case msg := <-c:
			handler(msg.Payload)
		}
	}
}

func subscribe(c <-chan *redis.Message, timeout time.Duration, handler func(msg string) int) int {
	if timeout != 0 {
		timer := time.NewTicker(timeout)
		for {
			select {
			case <-timer.C:
				if timeout != 0 {
					return -2
				}
			case msg := <-c:
				res := handler(msg.Payload)
				if res != 0 {
					return res
				}
			}
		}
	} else {
		for {
			msg := <-c
			res := handler(msg.Payload)
			if res != 0 {
				return res
			}
		}
	}
}

func publish(channel string, msg chan string, shutdown chan struct{}) {
	ctx := context.Background()
	for {
		select {
		case m := <-msg:
			rdsClusterClient.Publish(ctx, channel, m)
		case <-shutdown:
			return
		}
	}
}

func (redisOpt) CreateSingleNode(port int, desc string) {
	currentdir, _ := os.Getwd()
	p := strconv.Itoa(port)
	path := currentdir + "/app-redis-node/" + p
	if !dirandfileopt.ExistDir(path) {
		dirandfileopt.CreateDir(path)
		dirandfileopt.AppendToConfigFile(path+"/redis.conf", "bind 0.0.0.0\n")
		dirandfileopt.AppendToConfigFile(path+"/redis.conf", "daemonize yes\n")
		dirandfileopt.AppendToConfigFile(path+"/redis.conf", "protected-mode no\n")
		dirandfileopt.AppendToConfigFile(path+"/redis.conf", "databases 1\n")
		dirandfileopt.AppendToConfigFile(path+"/redis.conf", "port "+p+"\n")
		dirandfileopt.AppendToConfigFile(path+"/redis.conf", "dir "+path+"\n")
		dirandfileopt.AppendToConfigFile(path+"/redis.conf", "always-show-logo yes\n")
		// dirandfileopt.AppendToConfigFile(path+"/redis.conf", "logfile "+path+"/run.log\n")
		dirandfileopt.AppendToConfigFile(path+"/redis.conf", "loglevel verbose\n")
		formatprint.FormatPrint(formatprint.Yellow, "redis端口:"+p+"节点创建成功, 你可以启动它了")
		RedisProxy.StartSingleNode(port)
	}
	if Config.Stormi.Ip == "" {
		formatprint.FormatPrint(formatprint.Magenta, "未配置ip, 当前节点无法加入配置文件")
		return
	}
	dirandfileopt.AppendToConfigFile(currentdir+"/app.config", "<redis-node@"+Config.Stormi.Ip+":"+p+"@"+desc+">\n")
	formatprint.FormatPrint(formatprint.Yellow, "<redis-node@"+Config.Stormi.Ip+":"+p+"@"+desc+">已载入配置文件")

}

func (redisOpt) CreateClusterNode(port int, desc string) {
	if Config.Stormi.Ip == "" {
		formatprint.FormatPrint(formatprint.Magenta, "未配置ip, 当前集群节点无法创建")
		return
	}
	currentdir, _ := os.Getwd()
	p := strconv.Itoa(port)
	path := currentdir + "/app-redis-cluster/" + p
	if !dirandfileopt.ExistDir(path) {
		dirandfileopt.CreateDir(path)
		dirandfileopt.AppendToConfigFile(path+"/redis.conf", "bind 0.0.0.0\n")
		dirandfileopt.AppendToConfigFile(path+"/redis.conf", "daemonize yes\n")
		dirandfileopt.AppendToConfigFile(path+"/redis.conf", "protected-mode no\n")
		dirandfileopt.AppendToConfigFile(path+"/redis.conf", "databases 1\n")
		dirandfileopt.AppendToConfigFile(path+"/redis.conf", "port "+p+"\n")
		dirandfileopt.AppendToConfigFile(path+"/redis.conf", "dir "+path+"\n")
		dirandfileopt.AppendToConfigFile(path+"/redis.conf", "always-show-logo yes\n")
		// dirandfileopt.AppendToConfigFile(path+"/redis.conf", "logfile "+path+"/run.log\n")
		dirandfileopt.AppendToConfigFile(path+"/redis.conf", "loglevel verbose\n")
		dirandfileopt.AppendToConfigFile(path+"/redis.conf", "cluster-enabled yes\n")
		dirandfileopt.AppendToConfigFile(path+"/redis.conf", "cluster-node-timeout 5000\n")
		dirandfileopt.AppendToConfigFile(path+"/redis.conf", "cluster-config-file "+path+"/nodes.conf\n")
		dirandfileopt.AppendToConfigFile(path+"/redis.conf", "replica-announce-ip "+Config.Stormi.Ip+"\n")
		formatprint.FormatPrint(formatprint.Yellow, "redis端口:"+p+"集群节点创建成功, 你可以启动它了")
		RedisProxy.StartClusterNode(port)
	}

	dirandfileopt.AppendToConfigFile(currentdir+"/app.config", "<redis-nodes@"+Config.Stormi.Ip+":"+p+"@"+desc+">\n")
	formatprint.FormatPrint(formatprint.Yellow, "<redis-nodes@"+Config.Stormi.Ip+":"+p+"@"+desc+">已载入配置文件")
}

func (redisOpt) StartSingleNode(port int) {
	currentdir, _ := os.Getwd()
	p := strconv.Itoa(port)
	path := currentdir + "/app-redis-node/" + p
	if !dirandfileopt.ExistDir(path) {
		formatprint.FormatPrint(formatprint.Magenta, "redis端口:"+p+"节点不存在")
		return
	}
	go func() {
		os.Remove(path + "/run.log")
		s := runtime.GOOS
		if s == "windows" {
			ExecCommand("start redis-server " + path + "/redis.conf")
		} else {
			ExecCommand("nohup redis-server " + path + "/redis.conf >/dev/nul 2>&1 &")
		}
	}()
	time.Sleep(2 * time.Second)
}

func (redisOpt) StartClusterNode(port int) {
	currentdir, _ := os.Getwd()
	p := strconv.Itoa(port)
	path := currentdir + "/app-redis-cluster/" + p
	if !dirandfileopt.ExistDir(path) {
		formatprint.FormatPrint(formatprint.Magenta, "redis端口:"+p+"节点不存在")
		return
	}
	go func() {
		// os.Remove(path + "/run.log")
		s := runtime.GOOS
		if s == "windows" {
			ExecCommand("start redis-server " + path + "/redis.conf")
		} else {
			ExecCommand("nohup redis-server " + path + "/redis.conf >/dev/nul 2>&1 &")
		}
	}()
	time.Sleep(2 * time.Second)
	if ConfigMap != nil {
		if rand.Intn(2) == 0 {
			ExecCommand("echo yes | redis-cli --cluster add-node " + Config.Stormi.Ip + ":" + p + " " + ConfigMap["redis-nodes"][0][0])
		} else {
			ExecCommand("echo yes | redis-cli --cluster add-node " + Config.Stormi.Ip + ":" + p + " " + ConfigMap["redis-nodes"][0][0] + " --cluster-slave")

		}
	}
}

func (redisOpt) shutdownNode(port int) {
	res := sh("stormi redis-kill " + strconv.Itoa(port))
	if res == "1" {
		fmt.Println("redis:" + strconv.Itoa(port) + "节点关闭成功")
	} else if res == "-1" {
		fmt.Println("redis:" + strconv.Itoa(port) + "节点关闭失败，可能redis并未启动或者redis实例并未创建")
	} else {
		fmt.Println(res)
	}
}

func (redisOpt) CreateCluster(port1, port2, port3, port4, port5, port6 int) {
	if Config.Stormi.Redis.Nodes != "" {
		fmt.Println("配置文件中存在集群节点，无需创建新集群")
		return
	}
	go func() {
		RedisProxy.CreateClusterNode(port1, "redis集群节点端口:"+strconv.Itoa(port1))
		RedisProxy.StartClusterNode(port1)
	}()
	go func() {
		RedisProxy.CreateClusterNode(port2, "redis集群节点端口:"+strconv.Itoa(port2))
		RedisProxy.StartClusterNode(port2)
	}()
	go func() {
		RedisProxy.CreateClusterNode(port3, "redis集群节点端口:"+strconv.Itoa(port3))
		RedisProxy.StartClusterNode(port3)
	}()
	go func() {
		RedisProxy.CreateClusterNode(port4, "redis集群节点端口:"+strconv.Itoa(port4))
		RedisProxy.StartClusterNode(port4)
	}()
	go func() {
		RedisProxy.CreateClusterNode(port5, "redis集群节点端口:"+strconv.Itoa(port5))
		RedisProxy.StartClusterNode(port5)
	}()
	go func() {
		RedisProxy.CreateClusterNode(port6, "redis集群节点端口:"+strconv.Itoa(port6))
		RedisProxy.StartClusterNode(port6)
	}()
	nodes := Config.Stormi.Ip + ":" + strconv.Itoa(port1) + " "
	nodes = nodes + Config.Stormi.Ip + ":" + strconv.Itoa(port2) + " "
	nodes = nodes + Config.Stormi.Ip + ":" + strconv.Itoa(port3) + " "
	nodes = nodes + Config.Stormi.Ip + ":" + strconv.Itoa(port4) + " "
	nodes = nodes + Config.Stormi.Ip + ":" + strconv.Itoa(port5) + " "
	nodes = nodes + Config.Stormi.Ip + ":" + strconv.Itoa(port6) + " "
	time.Sleep(1 * time.Second)
	ExecCommand("echo yes | redis-cli --cluster create --cluster-replicas 1 " + nodes)
	currentdir, _ := os.Getwd()
	dirandfileopt.AppendToYaml(currentdir+"/app.yaml", []string{nodes})
}

func (redisOpt) UploadProto(name string) {
	filename := currentDir + "/server/rpcserver/protos/" + name + ".proto"
	fileContent, err := os.ReadFile(filename)
	if err != nil {
		fmt.Println("读取文件失败:", err)
		return
	}
	err = rdsClusterClient.Set(context.Background(), "stormi:protos:"+name, fileContent, 0).Err()
	if err != nil {
		fmt.Println("上传文件失败:", err)
		return
	}
	fmt.Println("文件上传成功")
}

func (redisOpt) DownLoadProto(name string) {
	content, _ := rdsClusterClient.Get(context.Background(), "stormi:protos:"+name).Result()
	if content != "" {
		path := currentDir + "/server/serverset/protos/" + name + ".proto"
		file, err := os.OpenFile(path, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, 0644)
		if err != nil {
			fmt.Println("创建文件失败：", err)
		}
		defer file.Close()
		_, err = file.WriteString(content)
		if err != nil {
			fmt.Println("写入文件失败：", err)
		} else {
			fmt.Println("文件下载成功，目录/server/servserset/protos")
			Create.ProtoProxy(name)
		}
	} else {
		fmt.Println(name, "proto文件未在云端")
	}
}

func (redisOpt) UploadQueue(name string) {
	filename := currentDir + "/server/nsqd/queues/" + name + "Queue/" + name + "Queue.go"
	fileContent, err := os.ReadFile(filename)
	if err != nil {
		fmt.Println("读取文件失败:", err)
		return
	}
	err = rdsClusterClient.Set(context.Background(), "stormi:queues:"+name, fileContent, 0).Err()
	if err != nil {
		fmt.Println("上传文件失败:", err)
		return
	}
	fmt.Println("文件上传成功")
}

func (redisOpt) DownLoadQueue(name string) {
	content, _ := rdsClusterClient.Get(context.Background(), "stormi:queues:"+name).Result()
	if content != "" {
		path := currentDir + "/server/nsqd/queues/" + name + "Queue/" + name + "Queue.go"
		file, err := os.OpenFile(path, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, 0644)
		if err != nil {
			fmt.Println("创建文件失败：", err)
		}
		defer file.Close()
		_, err = file.WriteString(content)
		if err != nil {
			fmt.Println("写入文件失败：", err)
		} else {
			fmt.Println("文件下载成功，目录/server/nsqd/queues/" + name + "Queue")
		}
	} else {
		fmt.Println(name, "队列文件未在云端")
	}
}
