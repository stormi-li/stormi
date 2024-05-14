# stormi-go 框架（Spring-Java 时代的终结者）

## 简介

stormi 框架所原创的代理方案集成了类似 Spring 的容器方案，Springboot 的自动配置方案和 Spring Cloud 的微服务方案的功能，并且更容易使用，功能更强大，可扩展性更强。stormi 原生集成了服务注册与发现，分布式锁，和分布式事务等功能，还提供了一套进程间（跨主机）通信的解决方案，让微服务的开发真正实现跨主机联动，使得多人同时开发同一个功能成为现实，基于该方案本框架实现了 Cooperation 代理用于替代传统的服务注册与发现模式，并且该代理在灵活性、可用性和易用性上优于当前所有的其它方案。该框架已经给各位实现了 Redis 代理，Config 代理，Server 代理，MySQL 代理，Transaction 代理，NSQD 代理和 Cooperation 代理，其中最强大的是 Redis 代理和 Config 代理，这两个代理是最底层的代理，所有的代理都依赖这两个代理，其他开发人员可以使用这两个代理开发自己的代理，同时我希望各位开发人员如果觉得该框架好用的话可以多多开发和开源自己的代理，让我们一起搭建比 Spring 生态更加强大的 stormi 生态。github 链接：https://github.com/stormi-li/stormi

## 使用教程

哔哩哔哩视频教程： https://www.bilibili.com/video/BV1Ft421c7uh/?share_source=copy_web&vd_source=1e94ae838a6ff53b4df3b5dbe78ba199

### 1. 安装

##### 安装 stormi 框架

```
go get -u github.com/stormi-li/stormi
```

##### 安装 redis-server, redis-cli, nsqd

```go
package main

import "github.com/stormi-li/stormi"

func main() {
	stormi.NodeBuilder.Install()
}
```

##### 启动 redis 单例，和 redis 集群

```go
//启动 redis 单例
package main

import "github.com/stormi-li/stormi"

func main() {
	stormi.NodeBuilder.CreateRedisNode(2131, stormi.NodeType.RedisStandalone, "127.0.0.1", "C:\\Users\\lilili\\Desktop\\stormistudy\\redisstandalone")
}
```

```go
//启动 Redis 集群
package main

import "github.com/stormi-li/stormi"

func main() {
	stormi.NodeBuilder.CreateRedisCluster(2131, 2132, 2133, 2134, 2135, 2136, "127.0.0.1", "C:\\Users\\lilili\\Desktop\\stormistudy\\rediscluster")
}
```

```go
//启动 Redis 集群节点并将其加入 Redis 集群当中
package main

import "github.com/stormi-li/stormi"

func main() {
	stormi.NodeBuilder.CreateRedisNode(2137, stormi.NodeType.RedisCluster, "127.0.0.1", "C:\\Users\\lilili\\Desktop\\stormistudy\\rediscluster")
	stormi.NodeBuilder.AddNodeToRedisCluster("127.0.0.1:2137", "127.0.0.1:2131", stormi.NodeType.RedisMaster)
	stormi.NodeBuilder.CreateRedisNode(2138, stormi.NodeType.RedisCluster, "127.0.0.1", "C:\\Users\\lilili\\Desktop\\stormistudy\\rediscluster")
	stormi.NodeBuilder.AddNodeToRedisCluster("127.0.0.1:2138", "127.0.0.1:2131", stormi.NodeType.RedisSlave)
}
```

##### 创建 Redis 代理连接集群并且查看集群信息

```go
package main

import "github.com/stormi-li/stormi"

func main() {
	rp := stormi.NewRedisProxy("127.0.0.1:2131")
	rp.RedisClusterNodesInfo()
}
----------------------------------
package main

import "github.com/stormi-li/stormi"

func main() {
	rp := stormi.NewRedisProxy("127.0.0.1:213")
	rp.RedisSingleNodeInfo()
}
```

### 2. Redis 代理的使用

##### 获取 redisClient 和 redisClusterClient

```go
package main

import (
	"context"

	"github.com/stormi-li/stormi"
)

func main() {
	rp := stormi.NewRedisProxy("127.0.0.1:213")
	rc := rp.RedisClient()
	// rcc := rp.RedisClusterClient()
	rc.Set(context.Background(), "stormi", "stormi", 0)
}
-----------------------------------------------------------
package main

import (
	"context"

	"github.com/stormi-li/stormi"
)

func main() {
	rp := stormi.NewRedisProxy("127.0.0.1:2131")
	//rp.RedisClient()
	rcc := rp.RedisClusterClient()
	rcc.Set(context.Background(), "stormi", "stormi", 0)
}
```

##### 分布式锁的使用

```go
//分布式锁，实现了看门狗机制和锁id识别
package main

import (
	"fmt"
	"time"

	"github.com/stormi-li/stormi"
)

func main() {
	rp := stormi.NewRedisProxy("127.0.0.1:2131")
	go func() {
		l := rp.NewLock("lock")
		for {
			l.Lock()
			fmt.Println(1)
			time.Sleep(1000 * time.Millisecond)
			l.UnLock()
			time.Sleep(20 * time.Millisecond)
		}
	}()
	go func() {
		l := rp.NewLock("lock")
		for {
			l.Lock()
			fmt.Println(2)
			l.UnLock()
		}
	}()
	select {}
}
```

##### Notify 和 Wait 的使用

```go
package main

import (
	"fmt"
	"time"

	"github.com/stormi-li/stormi"
)

func main() {
	rp := stormi.NewRedisProxy("127.0.0.1:2131")
	res := rp.Wait("stormi-channel", 100*time.Second)
	fmt.Println(res)
}
-----------------------------------------------------------
package main

import (
	"github.com/stormi-li/stormi"
)

func main() {
	rp := stormi.NewRedisProxy("127.0.0.1:2131")
	rp.Notify("stormi-channel", "hello world!!!")
}
```

##### CycleWait 的使用

```go
package main

import (
	"fmt"
	"time"

	"github.com/stormi-li/stormi"
)

func main() {
	rp := stormi.NewRedisProxy("127.0.0.1:2131")
	rp.CycleWait("stormi-channel", 5*time.Second, true, func(msg *string) error {
		if msg != nil {
			fmt.Println(*msg)
		} else {
			fmt.Println("超时")
		}
		return nil
	})
}
-----------------------------------------------------------
package main

import (
	"time"

	"github.com/stormi-li/stormi"
)

func main() {
	rp := stormi.NewRedisProxy("127.0.0.1:2131")
	for {
		rp.Notify("stormi-channel", "stormi-notify")
		time.Sleep(1000 * time.Millisecond)
	}
}
```

##### Publish 和 Subscribe 使用

```go
package main

import (
	"fmt"

	"github.com/stormi-li/stormi"
)

func main() {
	rp := stormi.NewRedisProxy("127.0.0.1:2131")
	pubsub := rp.GetPubSub("stormi-pubsub")

	//开启订阅
	rp.Subscribe(pubsub, 0, func(msg string) int {
		if msg == "stop" {
			return 1
		}
		fmt.Println(msg)
		return 0
	})
	fmt.Println("订阅关闭")
}
-----------------------------------------------------------
package main

import (
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/stormi-li/stormi"
)

func main() {
	rp := stormi.NewRedisProxy("127.0.0.1:2131")
	msgc := make(chan string, 1)
	sdc := make(chan struct{}, 1)
	exit := make(chan struct{}, 1)
	go func() {
		//开启发布
		rp.Publish("stormi-pubsub", msgc, sdc)
		fmt.Println("发布关闭")
		exit <- struct{}{}
	}()

	for i := 0; i < 10; i++ {
		msgc <- uuid.NewString()
		time.Sleep(1 * time.Second)
	}

	msgc <- "stop"
	sdc <- struct{}{}
	<-exit
}
```

##### 快速搭建聊天室

```go
package main

import (
	"fmt"

	"github.com/stormi-li/stormi"
)

func main() {
	rp := stormi.NewRedisProxy("127.0.0.1:2131")
	pubsub := rp.GetPubSub("stormi-chat")
	rp.Subscribe(pubsub, 0, func(msg string) int {
		fmt.Println(msg)
		return 0
	})
}
----------------------------------
package main

import (
	"bufio"
	"os"
	"strings"

	"github.com/stormi-li/stormi"
)

func main() {
	rp := stormi.NewRedisProxy("127.0.0.1:2131")
	reader := bufio.NewReader(os.Stdin)
	for {
		text, _ := reader.ReadString('\n')
		text = strings.TrimSpace(text)
		rp.Notify("stormi-chat", text)
	}
}
```

### 3. Config 代理的使用

##### 注册配置信息到 redis 配置集以及配置持久化

```go
package main

import (
	"github.com/stormi-li/stormi"
)

func main() {
	cp := stormi.NewConfigProxy(stormi.NewRedisProxy("127.0.0.1:2131"))
	c := cp.NewConfig()
	c.Name = "nsqd"
	c.Addr = "127.0.0.1:3131"
	cp.Register(c)
    cp.ConfigPersistence()
}
```

##### 拉取配置信息

```go
package main

import (
	"fmt"

	"github.com/stormi-li/stormi"
)

func main() {
	cp := stormi.NewConfigProxy(stormi.NewRedisProxy("127.0.0.1:2131"))
	cmap := cp.Pull("nsqd")
	for _, c := range cmap {
		fmt.Printf("c: %v\n", c)
	}
}
```

##### 修改配置信息

```go
package main

import (
	"github.com/stormi-li/stormi"
)

func main() {
	cp := stormi.NewConfigProxy(stormi.NewRedisProxy("127.0.0.1:2131"))
	cmap := cp.Pull("nsqd")
	for _, c := range cmap {
		c.Ignore = true
		cp.Update(c)
	}
}
```

##### 删除配置信息和注册信息

```go
package main

import (
	"github.com/stormi-li/stormi"
)

func main() {
	cp := stormi.NewConfigProxy(stormi.NewRedisProxy("127.0.0.1:2131"))
	cmap := cp.Pull("nsqd")
	for _, c := range cmap {
		c.Ignore = true
		cp.Remove(c)
	}
	cp.RemoveRegister("nsqd")
}
```

##### 批量刷新配置信息

```go
package main

import (
	"github.com/stormi-li/stormi"
)

func main() {
	cp := stormi.NewConfigProxy(stormi.NewRedisProxy("127.0.0.1:2131"))
	c1 := cp.NewConfig()
	c1.Name = "nsqd"
	c1.Addr = "127.0.0.1:3131"
	c2 := cp.NewConfig()
	c2.Name = "nsqd"
	c2.Addr = "127.0.0.1:3132"
    //如果配置不存在就会注册配置
	cp.Refreshs([]stormi.Config{c1, c2})
}
```

##### 批量删除配置信息

```go
package main

import (
	"github.com/stormi-li/stormi"
)

func main() {
	cp := stormi.NewConfigProxy(stormi.NewRedisProxy("127.0.0.1:2131"))
	cmap := cp.Pull("nsqd")
	cp.Removes(cmap)
}
```

##### Sync 同步配置

```go
package main

import (
	"time"

	"github.com/stormi-li/stormi"
)

func main() {
	cp := stormi.NewConfigProxy(stormi.NewRedisProxy("127.0.0.1:2131"))
	time.Sleep(3 * time.Second)
	cp.Sync()
}
```

##### NotifySync 通知同步配置

```go
package main

import "github.com/stormi-li/stormi"

func main() {
    stormi.NewConfigProxy(stormi.NewRedisProxy("127.0.0.1:2131"))
    select{}
}
-----------------------------------------------------------
package main

import "github.com/stormi-li/stormi"

func main() {
    cp := stormi.NewConfigProxy(stormi.NewRedisProxy("127.0.0.1:2131"))
	cp.NotifySync("同步信息")
}
```

##### AddConfigHandler 添加配置处理器

```go
package main

import (
	"fmt"

	"github.com/stormi-li/stormi"
)

func main() {
    cp := stormi.NewConfigProxy(stormi.NewRedisProxy("127.0.0.1:2131"))
	cp.AddConfigHandler("nsqd", func(cmap map[string]*stormi.Config) {
		for _, c := range cmap {
			fmt.Printf("c: %v\n", c)
		}
	})
}
```

##### AddConfigSyncNotificationHandler 添加配置同步消息处理器

```go
package main

import (
	"fmt"

	"github.com/stormi-li/stormi"
)

func main() {
    cp := stormi.NewConfigProxy(stormi.NewRedisProxy("127.0.0.1:2131"))
	cp.AddConfigSyncNotificationHandler(func(configProxy stormi.ConfigProxy, msg string) {
		fmt.Println(msg)
	})
    select {}
}
-----------------------------------------------------------
package main

import (
	"github.com/stormi-li/stormi"
)

func main() {
	cp := stormi.NewConfigProxy(stormi.NewRedisProxy("127.0.0.1:2131"))
	cp.NotifySync("同步新配置")
}
```

### 4. Server 代理的使用

##### 服务注册与发现

```go
package main

import (
	"time"

	"github.com/stormi-li/stormi"
)

func main() {
    cp := stormi.NewServerProxy(stormi.NewConfigProxy(stormi.NewRedisProxy("127.0.0.1:2131")))
	cp.ConfigProxy().AddConfigSyncNotificationHandler(func(configProxy stormi.ConfigProxy, msg string) {})
	cp.Register("stormiserver", "127.0.0.1:8888", 3, 3*time.Second)
	select {}
}
-----------------------------------------------------------
package main

import (
	"fmt"
	"time"

	"github.com/stormi-li/stormi"
)

func main() {
    cp := stormi.NewServerProxy(stormi.NewConfigProxy(stormi.NewRedisProxy("127.0.0.1:2131")))
	cp.Discover("stormiserver", 3*time.Second, func(addr string) error {
		fmt.Println(addr)
		return nil
	})
	select {}
}
```

### 5. MySQL 代理的使用

##### 注册 MySQL 配置信息

```go
package main

import (
	"github.com/stormi-li/stormi"
)

func main() {
    mp := stormi.NewMysqlProxy(stormi.NewConfigProxy(stormi.NewRedisProxy("127.0.0.1:2131")))
	mp.Register(33061, "192.168.37.139:3306", "root", "123456", "stormi")
}
```

##### 连接数据库

```go
package main

import (
	"github.com/stormi-li/stormi"
)

func main() {
    mp := stormi.NewMysqlProxy(stormi.NewConfigProxy(stormi.NewRedisProxy("127.0.0.1:2131")))
	mp.ConnectByNodeId(33061)
}
-----------------------------------------------------------------
//动态修改数据库连接
package main

import (
	"github.com/stormi-li/stormi"
)

func main() {
	cp := stormi.NewConfigProxy(stormi.NewRedisProxy("127.0.0.1:2131"))
	mp := stormi.NewMysqlProxy(cp)
	mp.ConnectByNodeId(33061)
	cp.AddConfigHandler("mysql", func(cmap map[string]*stormi.Config) {
		for _, c := range cmap {
			if c.NodeId == 33061 {
				mp.ConnectByConfig(*c)
				break
			}
		}
	})
    select {}
}
-----------------------------------------------------------------
package main

import (
	"github.com/stormi-li/stormi"
)

func main() {
	cp := stormi.NewConfigProxy(stormi.NewRedisProxy("127.0.0.1:2131"))
	cmap := cp.Pull("mysql")
	for _, c := range cmap {
		c.NodeId = 33061
		c.Addr = "192.168.37.139:3306"
		cp.Update(c)
		cp.NotifySync("33061节点数据库更新, 请重新连接数据库")
		break
	}
}
```

##### 创建 ConfigTable，并插入数据

```go
package main

import (
	"github.com/stormi-li/stormi"
)

func main() {
    mp := stormi.NewMysqlProxy(stormi.NewConfigProxy(stormi.NewRedisProxy("127.0.0.1:2131")))
	mp.ConnectByNodeId(33061)
	ct := ConfigTable{}
	mp.DB().AutoMigrate(&ct)
	ct.Name = "nsqd"
	ct.Addr = "127.0.0.1:3131"
	mp.DB().Create(&ct)
}

type ConfigTable struct {
	Name string
	Addr string
}
```

### 6. Transaction 代理的使用

##### 使用分布式事务

```go
package main

import (
	"fmt"
	"math/rand"
	"strconv"

	"github.com/stormi-li/stormi"
)

func main() {
	tp := stormi.NewTransactionProxy(stormi.NewRedisProxy("127.0.0.1:2131"))
	ids := tp.NewDTxIds(3)
	server1(ids[0])
	server2(ids[1])
	server3(ids[2])
	tp.DCommit(ids, func(statement [][2]string) {
		fmt.Println("分布式事务失败")
		fmt.Println(statement)
	})
	select {}
}

func server1(id string) {
	mp := stormi.NewMysqlProxy(stormi.NewConfigProxy(stormi.NewRedisProxy("127.0.0.1:2131")))
	mp.ConnectByNodeId(33061)
	ct := ConfigTable{}
	ct.Name = "nsqd"
	ct.Addr = "127.0.0.1:" + strconv.Itoa(rand.Intn(10000))
	dtx := mp.NewDTx(id)
	dtx.DB().Create(&ct)
	dtx.Rollback()
}
func server2(id string) {
	mp := stormi.NewMysqlProxy(stormi.NewConfigProxy(stormi.NewRedisProxy("127.0.0.1:2131")))
	mp.ConnectByNodeId(33061)
	ct := ConfigTable{}
	ct.Name = "nsqd"
	ct.Addr = "127.0.0.1:" + strconv.Itoa(rand.Intn(10000))
	dtx := mp.NewDTx(id)
	dtx.DB().Create(&ct)
	dtx.Commit()
}
func server3(id string) {
	mp := stormi.NewMysqlProxy(stormi.NewConfigProxy(stormi.NewRedisProxy("127.0.0.1:2131")))
	mp.ConnectByNodeId(33061)
	ct := ConfigTable{}
	ct.Name = "nsqd"
	ct.Addr = "127.0.0.1:" + strconv.Itoa(rand.Intn(10000))
	dtx := mp.NewDTx(id)
	dtx.DB().Create(&ct)
	dtx.Commit()
}

type ConfigTable struct {
	Name string
	Addr string
}
```

### 7. nsqd 代理的使用

##### 搭建 nsqd 集群

```go
package main

import (
	"github.com/stormi-li/stormi"
)

func main() {
    np := stormi.NewNsqdProxy(stormi.NewConfigProxy(stormi.NewRedisProxy("127.0.0.1:2131")))
	stormi.NodeBuilder.CreateNsqdNode(4441, 4442, "C:\\Users\\lilili\\Desktop\\nsqd")
	stormi.NodeBuilder.CreateNsqdNode(4443, 4444, "C:\\Users\\lilili\\Desktop\\nsqd")
	stormi.NodeBuilder.CreateNsqdNode(4445, 4446, "C:\\Users\\lilili\\Desktop\\nsqd")
	stormi.NodeBuilder.CreateNsqdNode(4447, 4448, "C:\\Users\\lilili\\Desktop\\nsqd")
	np.Register("127.0.0.1:4441")
	np.Register("127.0.0.1:4443")
	np.Register("127.0.0.1:4445")
	np.Register("127.0.0.1:4447")
}
--------------------------------------
//启动 60 个 nsqd 节点
package main

import (
	"strconv"

	"github.com/stormi-li/stormi"
)

func main() {
	np := stormi.NewNsqdProxy(stormi.NewConfigProxy(stormi.NewRedisProxy("127.0.0.1:2131")))
	for i := 0; i < 60; i++ {
		stormi.NodeBuilder.CreateNsqdNode(4440+2*i, 4440+2*i+1, "C:\\Users\\lilili\\Desktop\\nsqd")
		np.Register("127.0.0.1:" + strconv.Itoa(4440+2*i))
	}
}
```

##### 生产和消费

```go
package main

import (
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/nsqio/go-nsq"
	"github.com/stormi-li/stormi"
)

func main() {
    np := stormi.NewNsqdProxy(stormi.NewConfigProxy(stormi.NewRedisProxy("127.0.0.1:2131")))
	np.AddConsumeHandler("stormi-nsqd", "channel1", func(message *nsq.Message) error {
		fmt.Println(string(message.Body))
		return nil
	})
	for {
		np.Publish("stormi-nsqd", []byte(uuid.NewString()))
		time.Sleep(200 * time.Millisecond)
	}
}
```

### 8. Cooperation 代理的使用

##### 创建 UserServer 协作协议

```go
package main

import (
	"github.com/stormi-li/stormi"
)

func main() {
	cop := stormi.NewCooperationProxy(stormi.NewConfigProxy(stormi.NewRedisProxy("127.0.0.1:2131")), "UserServer")
	cop.CreateCoprotocol()
}
```

##### 编辑 UserServer 协作协议

```go
package UserServer

// 协议码
const (
	Insert = iota
	Update
)

// 数据传输结构体
type UserServerDto struct {
	Id       int
	UserName string
}
```

##### 将编辑后的 UserServer 协作协议上传

```go
package main

import (
	"github.com/stormi-li/stormi"
)

func main() {
	cop := stormi.NewCooperationProxy(stormi.NewConfigProxy(stormi.NewRedisProxy("127.0.0.1:2131")), "UserServer")
	cop.PushCoprotocol()
}
```

##### 编辑 UserServer 协作协议的服务端代码

```go
package main

import (
	"encoding/json"
	"time"

	"github.com/stormi-li/stormi"
	UserServer "github.com/stormi-li/stormi/coprotocol/UserServer"
)

func main() {
	handler1()
	handler2()
	select {}
}

func handler1() {
	cop := stormi.NewCooperationProxy(stormi.NewConfigProxy(stormi.NewRedisProxy("127.0.0.1:2131")), "UserServer")
	cophd := cop.NewHandler()
	cophd.SetBufferSize(10)
	cophd.SetConcurrency(2)
	cophd.Handle(UserServer.Insert, func(data []byte) any {
		user := UserServer.UserServerDto{}
		json.Unmarshal(data, &user)
		user.Id = 1
		user.UserName = "handler1"
		time.Sleep(100 * time.Millisecond)
		return user
	})
}
func handler2() {
	cop := stormi.NewCooperationProxy(stormi.NewConfigProxy(stormi.NewRedisProxy("127.0.0.1:2131")), "UserServer")
	cophd := cop.NewHandler()
	cophd.SetBufferSize(10)
	cophd.SetConcurrency(2)
	cophd.Handle(UserServer.Insert, func(data []byte) any {
		user := UserServer.UserServerDto{}
		json.Unmarshal(data, &user)
		user.Id = 2
		user.UserName = "handler2"
		time.Sleep(200 * time.Millisecond)
		return user
	})
}
```

##### 编辑 UserServer 协作协议客户端代码

```go
package main

import (
	"fmt"
	"time"

	"github.com/stormi-li/stormi"
	UserServer "github.com/stormi-li/stormi/coprotocol/UserServer"
)

func main() {
	cop := stormi.NewCooperationProxy(stormi.NewConfigProxy(stormi.NewRedisProxy("127.0.0.1:2131")), "UserServer")
	caller := cop.NewCaller()
	caller.SetTimeout(10 * time.Second)
	caller.SetConcurrency(100)
	for i := 0; i < 50; i++ {
		go func() {
			for {
				dto := UserServer.UserServerDto{}
				caller.Call(UserServer.Insert, UserServer.UserServerDto{Id: 1, UserName: "stormi"}, &dto)
				fmt.Println(dto)
				time.Sleep(1 * time.Second)
			}
		}()
	}
	select {}
}
```

### 9.其它小工具的使用

##### Timer 计时器的使用

```go
package main

import (
	"fmt"
	"time"

	"github.com/stormi-li/stormi"
)

func main() {
	t := stormi.Utils.NewTimer()
	time.Sleep(100 * time.Millisecond)
	fmt.Println(t.Stamp())
	time.Sleep(100 * time.Millisecond)
	fmt.Println(t.StampAndReset())
	time.Sleep(100 * time.Millisecond)
	fmt.Println(t.Stamp())
}
```

##### StormiChat 聊天室的使用

```go
package main

import (
	"github.com/stormi-li/stormi"
)

func main() {
	sc := stormi.NewStormiChat(stormi.NewRedisProxy("127.0.0.1:2131"))
	sc.StartSub()
}
-----------------------------------------------------------
package main

import (
	"github.com/stormi-li/stormi"
)

func main() {
	sc := stormi.NewStormiChat(stormi.NewRedisProxy("127.0.0.1:2131"))
	sc.StartPub()
}
```
