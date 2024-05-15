package stormi

import (
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strconv"
	"strings"
	"time"
)

type nodeBuilder struct{}

var NodeBuilder nodeBuilder

type nodeType struct {
	RedisStandalone int
	RedisCluster    int
	RedisMaster     int
	RedisSlave      int
}

var NodeType nodeType

func init() {
	NodeType = nodeType{}
	NodeType.RedisCluster = 1
	NodeType.RedisStandalone = 2
	NodeType.RedisMaster = 3
	NodeType.RedisSlave = 4
}

func (nodeBuilder) Install() {
	var binpath string
	_, filename, _, _ := runtime.Caller(0)
	s := runtime.GOOS
	if s == "windows" {
		binpath = filepath.Dir(filename) + "\\bin-windows"
		gopath := os.Getenv("GOPATH")
		FileOpt.copyAllFiles(binpath, gopath+"\\bin")
	} else {
		binpath = filepath.Dir(filename) + "/bin-linux"
		gopath := os.Getenv("GOPATH")
		FileOpt.copyAllFiles(binpath, gopath+"/bin")
	}

}

func (nodeBuilder) CreateNsqdNode(tcpPort int, httpPort int, path string) {
	p1 := strconv.Itoa(tcpPort)
	p2 := strconv.Itoa(httpPort)
	s := runtime.GOOS
	go func() {
		if s == "windows" {
			path = path + "\\nsqdnode" + strconv.Itoa(tcpPort)
			os.MkdirAll(path, 0755)
			ExecCommand("start nsqd --tcp-address=0.0.0.0:" + p1 + " --http-address=0.0.0.0:" + p2 + " --data-path=" + path)
		} else {
			path = path + "/nsqdnode" + strconv.Itoa(tcpPort)
			os.MkdirAll(path, 0755)
			ExecCommand("nohup nsqd --tcp-address=0.0.0.0:" + p1 + " --http-address=0.0.0.0:" + p2 + " --data-path=" + path + " >> run.log >/dev/nul 2>&1 &")
		}
	}()
	time.Sleep(100 * time.Millisecond)
}

func (nodeBuilder) CreateRedisNode(port int, nodeType int, ip string, path string) {
	if nodeType == NodeType.RedisCluster {
		path = path + "/clusternode" + strconv.Itoa(port)
	} else if nodeType == NodeType.RedisStandalone {
		path = path + "/redisnode" + strconv.Itoa(port)
	} else {
		StormiFmtPrintln(magenta, noredis, "类型错误")
		return
	}

	if nodeType == NodeType.RedisStandalone {
		NodeBuilder.createRedisNode(port, path)
	}
	if nodeType == NodeType.RedisCluster {
		if ip == "" {
			StormiFmtPrintln(magenta, noredis, "请设置redis集群节点ip")
			return
		}
		NodeBuilder.createRedisCluster(port, ip, path)
	}
}

func (n nodeBuilder) CreateRedisCluster(port1, port2, port3, port4, port5, port6 int, ip string, path string) {
	n.CreateRedisNode(port1, NodeType.RedisCluster, ip, path)
	n.CreateRedisNode(port2, NodeType.RedisCluster, ip, path)
	n.CreateRedisNode(port3, NodeType.RedisCluster, ip, path)
	n.CreateRedisNode(port4, NodeType.RedisCluster, ip, path)
	n.CreateRedisNode(port5, NodeType.RedisCluster, ip, path)
	n.CreateRedisNode(port6, NodeType.RedisCluster, ip, path)
	nodes := ip + ":" + strconv.Itoa(port1) + " "
	nodes = nodes + ip + ":" + strconv.Itoa(port2) + " "
	nodes = nodes + ip + ":" + strconv.Itoa(port3) + " "
	nodes = nodes + ip + ":" + strconv.Itoa(port4) + " "
	nodes = nodes + ip + ":" + strconv.Itoa(port5) + " "
	nodes = nodes + ip + ":" + strconv.Itoa(port6) + " "
	ExecCommand("echo yes | redis-cli --cluster create --cluster-replicas 1 " + nodes)
}

func (nodeBuilder) createRedisNode(port int, path string) {
	os.MkdirAll(path, 0755)
	FileOpt.TruncateFile(path + "/redis.conf")
	FileOpt.AppendToFile(path+"/redis.conf", "bind 0.0.0.0\n")
	FileOpt.AppendToFile(path+"/redis.conf", "daemonize yes\n")
	FileOpt.AppendToFile(path+"/redis.conf", "protected-mode no\n")
	FileOpt.AppendToFile(path+"/redis.conf", "databases 1\n")
	FileOpt.AppendToFile(path+"/redis.conf", "port "+strconv.Itoa(port)+"\n")
	FileOpt.AppendToFile(path+"/redis.conf", "dir "+path+"\n")
	FileOpt.AppendToFile(path+"/redis.conf", "always-show-logo yes\n")
	FileOpt.AppendToFile(path+"/redis.conf", "loglevel verbose\n")
	s := runtime.GOOS
	go func() {
		if s == "windows" {
			ExecCommand("start redis-server " + path + "/redis.conf")
		} else {
			ExecCommand("nohup redis-server " + path + "/redis.conf >> run.log >/dev/nul 2>&1 &")
		}
	}()
	time.Sleep(100 * time.Millisecond)
}

func (nodeBuilder) createRedisCluster(port int, ip string, path string) {
	os.MkdirAll(path, 0755)
	FileOpt.TruncateFile(path + "/redis.conf")
	FileOpt.AppendToFile(path+"/redis.conf", "bind 0.0.0.0\n")
	FileOpt.AppendToFile(path+"/redis.conf", "daemonize yes\n")
	FileOpt.AppendToFile(path+"/redis.conf", "protected-mode no\n")
	FileOpt.AppendToFile(path+"/redis.conf", "databases 1\n")
	FileOpt.AppendToFile(path+"/redis.conf", "port "+strconv.Itoa(port)+"\n")
	FileOpt.AppendToFile(path+"/redis.conf", "dir "+path+"\n")
	FileOpt.AppendToFile(path+"/redis.conf", "always-show-logo yes\n")
	FileOpt.AppendToFile(path+"/redis.conf", "loglevel verbose\n")
	FileOpt.AppendToFile(path+"/redis.conf", "cluster-enabled yes\n")
	FileOpt.AppendToFile(path+"/redis.conf", "cluster-node-timeout 5000\n")
	FileOpt.AppendToFile(path+"/redis.conf", "cluster-config-file "+path+"/nodes.conf\n")
	FileOpt.AppendToFile(path+"/redis.conf", "replica-announce-ip "+ip+"\n")
	s := runtime.GOOS
	go func() {
		if s == "windows" {
			ExecCommand("start redis-server " + path + "/redis.conf")
		} else {
			ExecCommand("nohup redis-server " + path + "/redis.conf >> run.log >/dev/nul 2>&1 &")
		}
	}()
	time.Sleep(100 * time.Millisecond)
}

func (nodeBuilder) AddNodeToRedisCluster(newaddr, clusteraddr string, t int) {
	go func() {
		if t == NodeType.RedisMaster {
			ExecCommand("echo yes | redis-cli --cluster add-node " + newaddr + " " + clusteraddr)
		} else if t == NodeType.RedisSlave {
			ExecCommand("echo yes | redis-cli --cluster add-node " + newaddr + " " + clusteraddr + " --cluster-slave")
		}
	}()
	time.Sleep(100 * time.Millisecond)
}

func ExecCommand(cmd string) {
	StormiFmtPrintln(green, noredis, "执行脚本:", cmd)
	os := runtime.GOOS

	if os == "windows" {
		cmd := strings.ReplaceAll(cmd, "/", "\\")
		exec.Command("cmd", "/C", cmd).CombinedOutput()
	} else {
		exec.Command("bash", "-c", cmd).CombinedOutput()
	}
}
