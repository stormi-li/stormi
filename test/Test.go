package main

import (
	"github.com/stormi-li/stormi"
)

func main() {
	mp := stormi.NewMysqlProxy("127.0.0.1:2131")
	mp.ConnectByNodeId(33061)
	ct := ConfigTable{}
	mp.DB().AutoMigrate(&ct)
	ct.Name = "nsqd"
	ct.Addr = "127.0.0.1:3131"
	mp.DB().Create(&ct)
}

type ConfigTable struct {
	Name string
	Addr string
}
