package stormi

import (
	"context"
	"time"

	"github.com/go-redis/redis/v8"
	"github.com/google/uuid"
)

type RedisProxy struct {
	addrs            []string
	rdsClient        *redis.Client
	rdsClusterClient *redis.ClusterClient
	isCluster        bool
	isConnected      bool
}

func NewRedisProxy(addr interface{}) *RedisProxy {
	p := RedisProxy{}
	a, ok := addr.(string)
	if ok {
		rdsC := redis.NewClient(&redis.Options{
			Addr: a,
		})
		_, err := rdsC.ClusterNodes(context.Background()).Result()
		if err == nil {
			rdsCC := redis.NewClusterClient(&redis.ClusterOptions{
				Addrs: []string{a},
			})
			res, _ := rdsCC.Set(context.Background(), "stormi", "stormi", 0).Result()
			if res != "" {
				p.isCluster = true
				p.isConnected = true
				StormiFmtPrintln(yellow, a, "成功连接到redis集群:", a)
				p.rdsClusterClient = rdsCC
				p.addrs = []string{a}
				return &p
			}
		}
		res, _ := rdsC.Set(context.Background(), "stormi", "stormi", 0).Result()
		if res != "" {
			p.isConnected = true
			StormiFmtPrintln(yellow, a, "成功连接到redis单例:", a)
			p.rdsClient = rdsC
			p.addrs = []string{a}
			return &p
		}
	}
	s, ok := addr.([]string)
	if ok {
		rdsCC := redis.NewClusterClient(&redis.ClusterOptions{
			Addrs: s,
		})
		p.isCluster = true
		res, _ := rdsCC.Set(context.Background(), "stormi", "stormi", 0).Result()
		if res != "" {
			p.isCluster = true
			p.isConnected = true
			StormiFmtPrintln(yellow, s[0], "成功连接到redis集群:", s)
			p.rdsClusterClient = rdsCC
			p.addrs = s
			return &p
		}
	}
	StormiFmtPrintln(magenta, "连接redis失败", addr)
	return nil
}

func (rp *RedisProxy) RedisClient(id int) *redis.Client {
	if !rp.isConnected {
		StormiFmtPrintln(red, "当前未连接到任何redis节点")
	}
	if rp.isCluster {
		StormiFmtPrintln(magenta, rp.addrs[0], "当前redis为集群模式, 建议使用redis集群")
	}
	if id == 0 {
		return rp.rdsClient
	}
	StormiFmtPrintln(magenta, rp.addrs[0], "无法在配置集里面找到该NodeId的节点, 已返回当前redis节点")
	return rp.rdsClient
}

func (rp *RedisProxy) RedisClusterClient() *redis.ClusterClient {
	if !rp.isConnected {
		StormiFmtPrintln(red, "当前未连接到任何redis节点")
	}
	if !rp.isCluster {
		StormiFmtPrintln(magenta, rp.addrs[0], "当前redis为单例模式, 无法使用redis集群")
	}
	return rp.rdsClusterClient
}

func (rp *RedisProxy) RedisSingleNodeInfo() {
	opt := rp.rdsClient.Options()
	StormiFmtPrintln(cyan, rp.addrs[0], "当前redis节点地址:"+opt.Addr)
}

func (rp *RedisProxy) RedisClusterNodesInfo() {
	redisNodes, _ := rp.rdsClusterClient.ClusterNodes(context.Background()).Result()
	StormiFmtPrintln(cyan, rp.addrs[0], "当前redis集群信息:\n"+redisNodes)
}

type DLock struct {
	uuid     string
	lockName string
	stop     chan struct{}
	rp       *RedisProxy
}

func (rp *RedisProxy) NewLock(lockName string) *DLock {
	dLock := DLock{}
	dLock.lockName = lockName
	dLock.uuid = uuid.New().String()
	dLock.stop = make(chan struct{})
	dLock.rp = rp
	return &dLock
}

func (l *DLock) Lock() {
	ctx := context.Background()
	for {
		var ok bool
		if l.rp.isCluster {
			ok, _ = l.rp.rdsClusterClient.SetNX(ctx, l.lockName, l.uuid, 3*time.Second).Result()
		} else {
			ok, _ = l.rp.rdsClient.SetNX(ctx, l.lockName, l.uuid, 3*time.Second).Result()
		}

		if ok {
			go func() {
				ticker := time.NewTicker(1 * time.Second)
				defer ticker.Stop()
				for {
					select {
					case <-ticker.C:
						if l.rp.isCluster {
							ok, _ = l.rp.rdsClusterClient.SetNX(ctx, l.lockName, l.uuid, 3*time.Second).Result()
						} else {
							ok, _ = l.rp.rdsClient.SetNX(ctx, l.lockName, l.uuid, 3*time.Second).Result()
						}
					case <-l.stop:
						return
					}
				}
			}()
			break
		} else {
			l.rp.Wait(l.lockName, 3*time.Second)
		}
	}
}

func (l *DLock) UnLock() {
	l.stop <- struct{}{}
	ctx := context.Background()
	var uuid string
	if l.rp.isCluster {
		uuid, _ = l.rp.rdsClusterClient.Get(ctx, l.lockName).Result()
	} else {
		uuid, _ = l.rp.rdsClient.Get(ctx, l.lockName).Result()
	}

	if uuid == l.uuid {
		if l.rp.isCluster {
			l.rp.rdsClusterClient.Del(ctx, l.lockName)
		} else {
			l.rp.rdsClient.Del(ctx, l.lockName)
		}
		l.rp.Notify(l.lockName, "unlock")
	}
}

func (rp *RedisProxy) Notify(channel, msg string) {
	if rp.isCluster {
		rp.rdsClusterClient.Publish(context.Background(), channel, msg)
	} else {
		rp.rdsClient.Publish(context.Background(), channel, msg)
	}
}

func (rp *RedisProxy) Wait(channel string, timeout time.Duration) string {
	var pubsub *redis.PubSub
	if rp.isCluster {
		pubsub = rp.rdsClusterClient.Subscribe(context.Background(), channel)
	} else {
		pubsub = rp.rdsClient.Subscribe(context.Background(), channel)
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

func (rp *RedisProxy) CycleWait(channel string, timeout time.Duration, handler func(msg string)) {
	t := timeout
	var pubsub *redis.PubSub
	if rp.isCluster {
		pubsub = rp.rdsClusterClient.Subscribe(context.Background(), channel)
	} else {
		pubsub = rp.rdsClient.Subscribe(context.Background(), channel)
	}
	defer pubsub.Close()
	c := pubsub.Channel()
	timer := time.NewTicker(t)
	for {
		select {
		case <-timer.C:
			timer = time.NewTicker(t)
			handler("")
		case msg := <-c:
			handler(msg.Payload)
		}
	}
}

func (rp *RedisProxy) GetSubChannel(c string) <-chan *redis.Message {
	if rp.isCluster {
		return rp.rdsClusterClient.Subscribe(context.Background(), c).Channel()
	} else {
		return rp.rdsClient.Subscribe(context.Background(), c).Channel()
	}
}

func (rp *RedisProxy) Subscribe(c <-chan *redis.Message, timeout time.Duration, handler func(msg string) int) int {
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

func (rp *RedisProxy) Publish(channel string, msg chan string, shutdown chan struct{}) {
	ctx := context.Background()
	for {
		select {
		case m := <-msg:
			if rp.isCluster {
				rp.rdsClusterClient.Publish(ctx, channel, m)
			} else {
				rp.rdsClient.Publish(ctx, channel, m)
			}
		case <-shutdown:
			return
		}
	}
}
