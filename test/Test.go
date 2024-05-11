package main

import (
	"github.com/stormi-li/stormi"
)

func main() {
	stormi.NewCooperationProxy(stormi.NewConfigProxy(stormi.NewRedisProxy("127.0.0.1:2131")), "OrderServer")

}
