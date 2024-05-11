package main

import (
	"encoding/json"
	"fmt"
	"time"

	"github.com/stormi-li/stormi"
	OrderServer "github.com/stormi-li/stormi/coprotocol/OrderServer"
)

func main() {
	caller()
	select {}
}

func handler() {
	cop := stormi.NewCooperationProxy(stormi.NewConfigProxy(stormi.NewRedisProxy("127.0.0.1:2131")), "OrderServer")
	hd := cop.NewHandler()
	hd.Handle(OrderServer.Func1, func(data []byte) any {
		dto := OrderServer.OrderServerDto{}
		json.Unmarshal(data, &dto)
		fmt.Println(dto, "dddddddd")
		dto.Code = 10
		return dto
	})
}

func caller() {
	cop := stormi.NewCooperationProxy(stormi.NewConfigProxy(stormi.NewRedisProxy("127.0.0.1:2131")), "OrderServer")
	caller := cop.NewCaller()
	dto := OrderServer.OrderServerDto{}
	for {
		caller.Call(OrderServer.Func1, OrderServer.OrderServerDto{Code: 1, Message: "hi"}, &dto)
		fmt.Println(dto, "000000000000")
		time.Sleep(1 * time.Second)
	}
}
