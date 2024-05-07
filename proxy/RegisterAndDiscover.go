package proxy

import (
	"context"
	"fmt"
	"net"
	"strconv"
	"strings"
	"sync"
	"time"
)

type serverProxy struct{}

var ServerProxy = serverProxy{}

func (serverProxy) RegisterWithName(serverName string) {
	fmt.Println(serverName + "-rpc服务已注册")
	go func() {
		register(serverName)
	}()
}

func register(serverName string) {
	registerName := "stormi:serverset:" + serverName
	port := Config.Stormi.RpcServer.Port
	weight := Config.Stormi.RpcServer.Weight
	addr := Config.Stormi.Ip
	go func() {
		for {
			rdsClusterClient.Set(context.Background(), registerName, addr+":"+port+"-"+weight, 33*time.Second)
			time.Sleep(30 * time.Second)
		}
	}()
	go func() {
		for {
			notify("stormi:serverset:register-notify", serverName+"-"+addr+":"+port+"-"+weight)
			time.Sleep(3 * time.Second)
		}
	}()
}

var serverSet = make(map[string]chan string)

func discover() {
	sub := rdsClusterClient.Subscribe(context.Background(), "stormi:serverset:register-notify")
	go func() {
		subscribe(sub.Channel(), 0, func(msg string) int {
			parts := strings.Split(msg, "-")
			if serverSet[parts[0]] == nil {
				serverSet[parts[0]] = make(chan string, 10)
			}
			serverlist := serverSet[parts[0]]
			weight, _ := strconv.Atoi(parts[2])
			serverListLock.Lock()
			for i := 0; i < weight; i++ {
				if len(serverlist) == 10 {
					<-serverlist
				} else {
					serverlist <- parts[1]
				}
			}
			serverListLock.Unlock()
			return 0
		})
	}()
}

var serverListLock sync.Mutex

func getServerAddr(serverName string) string {
	serverlist := serverSet[serverName]
	if len(serverlist) == 0 {
		for {
			fmt.Println("searching", serverName, "rpc server......")
			time.Sleep(1 * time.Second)
			serverlist = serverSet[serverName]
			if len(serverlist) != 0 {
				break
			}
		}
	}
	addr := <-serverlist
	for {
		conn, err := net.DialTimeout("tcp", addr, 1*time.Second)
		conn.Close()
		if err == nil {
			break
		}
		fmt.Println(serverName, "rpc server addr:"+addr+" disconnected")
		fmt.Println("researching", serverName, "rpc server......")
		addr = <-serverlist
	}
	serverListLock.Lock()
	if len(serverlist) < 10 {
		serverlist <- addr
	}
	serverListLock.Unlock()
	return addr
}
