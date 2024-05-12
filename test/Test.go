package main

import (
	"fmt"
	"time"

	"github.com/stormi-li/stormi"
	OrderServer "github.com/stormi-li/stormi/coprotocol/OrderServer"
)

var cop = stormi.NewCooperationProxy(stormi.NewConfigProxy(stormi.NewRedisProxy("127.0.0.1:2131")), "OrderServer")
var caller = cop.NewCaller()

func main() {
	caller.SetTimeout(20 * time.Second)

	for i := 0; i < 20; i++ {
		go caller1()
	}
	select {}
}

func caller1() {
	for {
		dto := OrderServer.OrderServerDto{}
		caller.Call(OrderServer.Func1, OrderServer.OrderServerDto{Code: 1, Message: "hi"}, &dto)
		fmt.Println(dto)
	}
}
