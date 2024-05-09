package proxy

import (
	"fmt"
	"math/rand"
	"os"
	"runtime"
	"strconv"
	"time"

	"github.com/nsqio/go-nsq"
	"github.com/stormi-li/stormi/dirandfileopt"
	"github.com/stormi-li/stormi/formatprint"
)

var nsqConnPool = nsqConnectPool{}

type nsqdProxy struct{}

var NsqdProxy nsqdProxy

var nsqdConfig = nsq.NewConfig()

func nsqdInit() {
	nsqNodes := ConfigMap["nsqd-nodes"]
	for _, node := range nsqNodes {
		c, _ := nsq.NewProducer(node[0], nsqdConfig)
		c.SetLoggerLevel(nsq.LogLevelError)
		nsqConnPool.producers = append(nsqConnPool.producers, c)
		nsqConnPool.nodes = append(nsqConnPool.nodes, node[0])
	}
}

func (nsqdProxy) NsqClusterNodesDial() {
	for _, node := range nsqConnPool.nodes {
		ok := NsqdProxy.NsqNodeDial(node)
		if !ok {
			fmt.Println("nsqd", node, "connection refused")
		} else {
			fmt.Println("nsqd", node, "connected")
		}
	}
}

func (nsqdProxy) NsqNodeDial(node string) bool {
	c, _ := nsq.NewConsumer("connect-test", "test", nsqdConfig)
	c.SetLoggerLevel(nsq.LogLevelError)
	c.AddHandler(nsq.HandlerFunc(func(message *nsq.Message) error { return nil }))
	err := c.ConnectToNSQD(node)
	if err != nil {
		return false
	} else {
		return true
	}
}

type nsqConnectPool struct {
	producers []*nsq.Producer
	nodes     []string
}

func (nsqdProxy) NsqdConn() *nsqConnectPool {
	return &nsqConnPool
}

func (nsqdProxy) Publish(topic string, msg []byte) {
	for {
		randnum := rand.Intn(len(nsqConnPool.nodes))
		t := Utils.NewTicker()
		err := nsqConnPool.producers[randnum].Publish(topic, msg)
		t.Stamp(strconv.Itoa(randnum))
		if err == nil {
			break
		}
	}
}

func (nsqdProxy) ListenAndConsume(topic, channel string, hanlder func(message *nsq.Message) error) {
	config := nsq.NewConfig()
	for _, node := range nsqConnPool.nodes {
		c, _ := nsq.NewConsumer(topic, channel, config)
		c.AddHandler(nsq.HandlerFunc(hanlder))
		err := c.ConnectToNSQD(node)
		if err != nil {
			fmt.Println(err)
		}
	}
}

func (nsqdProxy) CreateNode(port int, desc string) {
	currentdir, _ := os.Getwd()
	p1 := strconv.Itoa(port)
	p2 := strconv.Itoa(port + 1)
	path := currentdir + "/app-nsqd-cluster/" + p1 + "-" + p2
	if !dirandfileopt.ExistDir(path) {
		dirandfileopt.CreateDir(path)
		formatprint.FormatPrint(formatprint.Yellow, "nsqd端口:"+p1+"节点创建成功,占用端口 "+p1+" "+p2+", 你可以启动它了")
		NsqdProxy.StartNode(port)
	}
	if Config.Stormi.Ip == "" {
		formatprint.FormatPrint(formatprint.Magenta, "未配置ip, 当前节点无法加入配置文件")
		return
	}
	dirandfileopt.AppendToConfigFile(currentdir+"/app.config", "<nsqd-nodes@"+Config.Stormi.Ip+":"+p1+"@"+desc+">\n")
	formatprint.FormatPrint(formatprint.Yellow, "<nsqd-nodes@"+Config.Stormi.Ip+":"+p1+"@"+desc+">已载入配置文件")
}

func (nsqdProxy) StartNode(port int) {
	currentdir, _ := os.Getwd()
	p1 := strconv.Itoa(port)
	p2 := strconv.Itoa(port + 1)
	path := currentdir + "/app-nsqd-cluster/" + p1 + "-" + p2
	go func() {
		s := runtime.GOOS
		if s == "windows" {
			ExecCommand("start nsqd -tcp-address=0.0.0.0:" + p1 + " -http-address=0.0.0.0:" + p2 + " -data-path=" + path)
		} else {
			ExecCommand("nohup nsqd -tcp-address=0.0.0.0:" + p1 + " -http-address=0.0.0.0:" + p2 + " -data-path=" + path + ">/dev/nul 2>&1 &")
		}
		//nsqd -tcp-address=0.0.0.0:5555 -http-address=0.0.0.0:5557
	}()
	time.Sleep(time.Second * 2)
}

func (nsqdProxy) shutdownNode(port int) {
	res := sh("stormi nsqd-kill " + strconv.Itoa(port))
	if res == "-1" {
		fmt.Println("关闭nsqd节点失败，可能是该节点已经被关闭")
	} else if res == "1" {
		fmt.Println("nsqd节点关闭成功")
	} else {
		fmt.Println(res)
	}
}
