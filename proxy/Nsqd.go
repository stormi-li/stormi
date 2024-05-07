package proxy

import (
	"fmt"
	"math/rand"
	"strconv"

	"github.com/nsqio/go-nsq"
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

var dir = sh("stormi-serversetdir.sh")

func (nsqdProxy) CreateNode(tport, hport int, desc string) {
	res := sh("stormi nsqd-create " + dir + " " + strconv.Itoa(tport) + " " + strconv.Itoa(hport))
	if res == "-1" {
		fmt.Println("创建nsqd节点失败，端口已被占用")
	} else if res == "1" {
		fmt.Println("nsqd:" + strconv.Itoa(tport) + "节点创建成功")
		node := Config.Stormi.Ip + ":" + strconv.Itoa(tport)
		s := "\n<nsqd-nodes@" + node + "@" + desc + ">"
		appendToConfigFile(s)
	} else {
		fmt.Println(res)
	}
}

func (nsqdProxy) StartNode(port int) {
	res := sh("stormi nsqd-start " + dir + " " + strconv.Itoa(port))
	if res == "-1" {
		fmt.Println("启动nsqd节点失败，可能是该节点并未被创建")
	} else if res == "-2" {
		fmt.Println("启动nsqd节点失败，可能是该节点已经被启动")
	} else if res == "1" {
		fmt.Println("nsqd节点启动成功，你可以查看run.log检查nsqd运行情况")
	} else {
		fmt.Println(res)
	}
}

func (nsqdProxy) ShutdownNode(port int) {
	res := sh("stormi nsqd-kill " + strconv.Itoa(port))
	if res == "-1" {
		fmt.Println("关闭nsqd节点失败，可能是该节点已经被关闭")
	} else if res == "1" {
		fmt.Println("nsqd节点关闭成功")
	} else {
		fmt.Println(res)
	}
}
