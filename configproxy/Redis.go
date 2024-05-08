package configproxy

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/go-redis/redis/v8"
	"github.com/google/uuid"
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
