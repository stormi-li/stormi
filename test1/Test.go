package main

import (
	"time"

	"github.com/stormi-li/stormi"
)

func main() {
	syp := stormi.NewSyncProxy(stormi.NewRedisProxy("127.0.0.1:213"))
	cond := syp.NewCond("cond")
	cond.Singal()
	time.Sleep(2 * time.Second)
	cond.Broadcast()
}
