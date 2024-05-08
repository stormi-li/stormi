package stormi

import (
	"context"
	"time"

	"github.com/go-redis/redis/v8"
	"github.com/google/uuid"
)

var rdsClient *redis.Client
var rdsClusterClient *redis.ClusterClient
var isCluster bool
var isConnected bool

type RedisOpt struct{}

var RedisProxy RedisOpt

func redisInit(addr interface{}) bool {
	a, ok := addr.(string)
	if ok {
		rdsClient = redis.NewClient(&redis.Options{
			Addr: a,
		})
		_, err := rdsClient.ClusterNodes(context.Background()).Result()
		if err == nil {
			rdsClusterClient = redis.NewClusterClient(&redis.ClusterOptions{
				Addrs: []string{a},
			})
			res, _ := rdsClusterClient.Set(context.Background(), "stormi", "stormi", 0).Result()
			if res != "" {
				isCluster = true
				isConnected = true
				StormiFmtPrintln(yellow, "成功连接到redis集群:", a)
				return true
			}
		}
		res, _ := rdsClient.Set(context.Background(), "stormi", "stormi", 0).Result()
		if res != "" {
			isConnected = true
			StormiFmtPrintln(yellow, "成功连接到redis单例:", a)
			return true
		}
	}
	s, ok := addr.([]string)
	if ok {
		rdsClusterClient = redis.NewClusterClient(&redis.ClusterOptions{
			Addrs: s,
		})
		isCluster = true
		res, _ := rdsClusterClient.Set(context.Background(), "stormi", "stormi", 0).Result()
		if res != "" {
			isCluster = true
			isConnected = true
			StormiFmtPrintln(yellow, "成功连接到redis集群:", s)
			return true
		}
	}
	StormiFmtPrintln(magenta, "连接redis失败", addr)
	return false
}

func (RedisOpt) RedisClient(id int) *redis.Client {
	if isCluster {
		StormiPrintln(magenta, "当前redis为集群模式, 建议使用redis集群")
	}
	if id == 0 {
		return rdsClient
	}
	cs := ConfigMap["redis-single"]
	for _, c := range cs {
		if c.NodeId == id {
			return redis.NewClient(&redis.Options{
				Addr: c.Addr,
			})
		}
	}
	StormiPrint(magenta, "无法在配置集里面找到该NodeId的节点, 已返回当前redis节点")
	return rdsClient
}

func (RedisOpt) RedisClusterClient() *redis.ClusterClient {
	if !isCluster {
		StormiPrintln(magenta, "当前redis为单例模式, 无法使用redis集群")
	}
	return rdsClusterClient
}

func (RedisOpt) RedisSingleNodeInfo() {
	opt := rdsClient.Options()
	StormiPrintln(cyan, "当前redis节点地址:"+opt.Addr)
}

func (RedisOpt) RedisClusterNodesInfo() {
	redisNodes, _ := rdsClusterClient.ClusterNodes(context.Background()).Result()
	StormiPrintln(cyan, "当前redis集群信息:\n"+redisNodes)
}

type DLock struct {
	uuid     string
	lockName string
	stop     chan struct{}
}

func (RedisOpt) NewLock(lockName string) *DLock {
	dLock := DLock{}
	dLock.lockName = lockName
	dLock.uuid = uuid.New().String()
	dLock.stop = make(chan struct{})
	return &dLock
}

func (l *DLock) Lock() {
	ctx := context.Background()
	for {
		var ok bool
		if isCluster {
			ok, _ = rdsClusterClient.SetNX(ctx, l.lockName, l.uuid, 3*time.Second).Result()
		} else {
			ok, _ = rdsClient.SetNX(ctx, l.lockName, l.uuid, 3*time.Second).Result()
		}

		if ok {
			go func() {
				ticker := time.NewTicker(1 * time.Second)
				defer ticker.Stop()
				for {
					select {
					case <-ticker.C:
						if isCluster {
							ok, _ = rdsClusterClient.SetNX(ctx, l.lockName, l.uuid, 3*time.Second).Result()
						} else {
							ok, _ = rdsClient.SetNX(ctx, l.lockName, l.uuid, 3*time.Second).Result()
						}
					case <-l.stop:
						return
					}
				}
			}()
			break
		} else {
			RedisProxy.Wait(l.lockName, 3*time.Second)
		}
	}
}

func (l *DLock) UnLock() {
	l.stop <- struct{}{}
	ctx := context.Background()
	var uuid string
	if isCluster {
		uuid, _ = rdsClusterClient.Get(ctx, l.lockName).Result()
	} else {
		uuid, _ = rdsClient.Get(ctx, l.lockName).Result()
	}

	if uuid == l.uuid {
		if isCluster {
			rdsClusterClient.Del(ctx, l.lockName)
		} else {
			rdsClient.Del(ctx, l.lockName)
		}
		RedisProxy.Notify(l.lockName, "unlock")
	}
}

func (RedisOpt) Notify(channel, msg string) {
	if isCluster {
		rdsClusterClient.Publish(context.Background(), channel, msg)
	} else {
		rdsClient.Publish(context.Background(), channel, msg)
	}
}

func (RedisOpt) Wait(channel string, timeout time.Duration) string {
	var pubsub *redis.PubSub
	if isCluster {
		pubsub = rdsClusterClient.Subscribe(context.Background(), channel)
	} else {
		pubsub = rdsClient.Subscribe(context.Background(), channel)
	}
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

func (RedisOpt) CycleWait(channel string, timeout time.Duration, handler func(msg string)) {
	var pubsub *redis.PubSub
	if isCluster {
		pubsub = rdsClusterClient.Subscribe(context.Background(), channel)
	} else {
		pubsub = rdsClient.Subscribe(context.Background(), channel)
	}
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

func (RedisOpt) Subscribe(c <-chan *redis.Message, timeout time.Duration, handler func(msg string) int) int {
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

func (RedisOpt) Publish(channel string, msg chan string, shutdown chan struct{}) {
	ctx := context.Background()
	for {
		select {
		case m := <-msg:
			if isCluster {
				rdsClusterClient.Publish(ctx, channel, m)
			} else {
				rdsClient.Publish(ctx, channel, m)
			}
		case <-shutdown:
			return
		}
	}
}
