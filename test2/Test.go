package main

import "github.com/stormi-li/stormi"

func main() {
	rp := stormi.NewRedisProxy("127.0.0.1:213")
	rp.RedisSingleNodeInfo()
}
