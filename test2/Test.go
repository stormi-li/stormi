package main

import (
	"fmt"
	"time"

	"github.com/stormi-li/stormi"
	UserServer "github.com/stormi-li/stormi/coprotocol/UserServer"
)

func main() {
	cop := stormi.NewCooperationProxy(stormi.NewConfigProxy(stormi.NewRedisProxy("127.0.0.1:2131")), "UserServer")
	caller := cop.NewCaller()
	caller.SetTimeout(10 * time.Second)
	caller.SetConcurrency(100)
	for i := 0; i < 50; i++ {
		go func() {
			for {
				dto := UserServer.UserServerDto{}
				caller.Call(UserServer.Insert, UserServer.UserServerDto{Id: 1, UserName: "stormi"}, &dto)
				fmt.Println(dto)
				time.Sleep(1 * time.Second)
			}
		}()
	}
	select {}
}
