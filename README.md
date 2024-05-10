# stormi-go框架（spring-java时代的终结者）

stormi框架所原创的代理方案集成了类似spring的容器方案，springboot的自动配置方案和springcloud的微服务方案的功能，并且更容易使用，功能更强大，可扩展性更强。stormi原生集成了服务注册与发现，分布式锁，和分布式事务等功能，还提供了一套进程间（跨主机）通信的解决方案，让软件的开发真正实现跨主机联动，使得多人同时开发同一个功能成为现实。该框架已经给各位实现了redis代理，config代理，server代理，mysql代理，transaction代理，nsqd代理，其中最强大的是redis代理和config代理，这两个代理是最底层的代理，所有的代理都依赖这两个代理，其他开发人员可以使用这两个代理开发自己的代理，同时我希望各位开发人员如果觉得该框架好用的话可以开发和开源自己的代理，让我们一起搭建比spring生态更加强大的stormi生态





## 使用教程

### 1.安装

- ##### 安装stormi框架

```
go get -u github.com/stormi-li/stormi
```



- ##### 安装redis-server,redis-cli,nsqd

```go
package main

import "github.com/stormi-li/stormi"

func main() {
	stormi.NodeBuilder.Install()
}
```

结果

```
[stormi-redis:0.0.0.0:0]: [已安装: nsqd.exe 到 $GOPATH\bin] 
[stormi-redis:0.0.0.0:0]: [已安装: redis-cli.exe 到 $GOPATH\bin] 
[stormi-redis:0.0.0.0:0]: [已安装: redis-server.exe 到 $GOPATH\bin]
```



- ##### 启动redis单例，和redis集群

```go
//启动redis单例
package main

import "github.com/stormi-li/stormi"

func main() {
	stormi.NodeBuilder.CreateRedisNode(213, stormi.NodeType.RedisStandalone, "127.0.0.1", "C:\\Users\\lilili\\Desktop\\stormistudy\\redisstandalone")
}
```

```go
//启动redis集群
package main

import "github.com/stormi-li/stormi"

func main() {
	stormi.NodeBuilder.CreateRedisCluster(2131, 2132, 2133, 2134, 2135, 2136, "127.0.0.1", "C:\\Users\\lilili\\Desktop\\stormistudy\\rediscluster")
}

```

```go
//启动redis集群节点并将其加入当redis集群当中
package main

import "github.com/stormi-li/stormi"

func main() {
	stormi.NodeBuilder.CreateRedisNode(2137, stormi.NodeType.RedisCluster, "127.0.0.1", "C:\\Users\\lilili\\Desktop\\stormistudy\\rediscluster")
	stormi.NodeBuilder.AddNodeToRedisCluster("127.0.0.1:2137", "127.0.0.1:2131", stormi.NodeType.RedisMaster)
	stormi.NodeBuilder.CreateRedisNode(2138, stormi.NodeType.RedisCluster, "127.0.0.1", "C:\\Users\\lilili\\Desktop\\stormistudy\\rediscluster")
	stormi.NodeBuilder.AddNodeToRedisCluster("127.0.0.1:2138", "127.0.0.1:2131", stormi.NodeType.RedisSlave)
}
```



- ##### 创建redis代理连接集群并且查看集群信息

```go
package main

import "github.com/stormi-li/stormi"

func main() {
	rp := stormi.NewRedisProxy("127.0.0.1:2131")
	rp.RedisClusterNodesInfo()
}
```



### 2.redis代理使用

