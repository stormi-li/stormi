package stormi

import (
	"github.com/stormi-li/stormi/dirandfileopt"
	"github.com/stormi-li/stormi/proxy"
)

var RedisProxy = proxy.RedisProxy

func Init() {
	dirandfileopt.CreateDir("app-redis-cluster")
	dirandfileopt.CreateDir("app-nsqd-cluster")
	dirandfileopt.CreateDir("app-redis-node")
}
