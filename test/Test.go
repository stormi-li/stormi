package main

import (
	"encoding/json"
	"time"

	"github.com/stormi-li/stormi"
	UserServer "github.com/stormi-li/stormi/coprotocol/UserServer"
)

func main() {
	handler1()
	handler2()
	select {}
}

func handler1() {
	cop := stormi.NewCooperationProxy(stormi.NewConfigProxy(stormi.NewRedisProxy("127.0.0.1:2131")), "UserServer")
	cophd := cop.NewHandler()
	cophd.SetBufferSize(10)
	cophd.SetConcurrency(2)
	cophd.Handle(UserServer.Insert, func(data []byte) any {
		user := UserServer.UserServerDto{}
		json.Unmarshal(data, &user)
		user.Id = 1
		user.UserName = "handler1"
		time.Sleep(100 * time.Millisecond)
		return user
	})
}
func handler2() {
	cop := stormi.NewCooperationProxy(stormi.NewConfigProxy(stormi.NewRedisProxy("127.0.0.1:2131")), "UserServer")
	cophd := cop.NewHandler()
	cophd.SetBufferSize(10)
	cophd.SetConcurrency(2)
	cophd.Handle(UserServer.Insert, func(data []byte) any {
		user := UserServer.UserServerDto{}
		json.Unmarshal(data, &user)
		user.Id = 2
		user.UserName = "handler2"
		time.Sleep(200 * time.Millisecond)
		return user
	})
}
