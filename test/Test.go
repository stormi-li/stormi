package main

import (
	"time"

	"github.com/stormi-li/stormi"
)

func main() {
	np := stormi.NewNsqdProxy(stormi.NewConfigProxy(stormi.NewRedisProxy("127.0.0.1:2131")))
	time.Sleep(1 * time.Second)
	np.Publish("nsqd", []byte("message"))
	select {}
}
