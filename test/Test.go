package main

import (
	"encoding/json"
	"fmt"

	"github.com/stormi-li/stormi"
	OrderServer "github.com/stormi-li/stormi/coprotocol/OrderServer"
)

func main() {
	// handler()

	for i := 0; i < 13; i++ {
		go caller1()
	}
	select {}
}

func handler() {
	cop := stormi.NewCooperationProxy(stormi.NewConfigProxy(stormi.NewRedisProxy("127.0.0.1:2131")), "OrderServer")
	hd := cop.NewHandler()
	hd.Handle(OrderServer.Func1, func(data []byte) any {
		dto := OrderServer.OrderServerDto{}
		json.Unmarshal(data, &dto)
		fmt.Println(dto)
		dto.Code = 10
		return dto
	})
}

var cop = stormi.NewCooperationProxy(stormi.NewConfigProxy(stormi.NewRedisProxy("127.0.0.1:2131")), "OrderServer")
var caller = cop.NewCaller()

func caller1() {
	for {
		dto := OrderServer.OrderServerDto{}
		caller.Call(OrderServer.Func1, OrderServer.OrderServerDto{Code: 1, Message: "hi"}, &dto)
		fmt.Println(dto)
	}
}
