package main

import (
	"fmt"
	"math/rand"
	"strconv"

	"github.com/stormi-li/stormi"
)

func main() {
	tp := stormi.NewTransactionProxy(stormi.NewRedisProxy("127.0.0.1:2131"))
	ids := tp.NewDTxIds(3)
	// server1(ids[0])
	server2(ids[1])
	server3(ids[2])
	tp.DCommit(ids, func(statement [][2]string) {
		fmt.Println("分布式事务失败")
		fmt.Println(statement)
	})
	select {}
}

func server1(id string) {
	mp := stormi.NewMysqlProxy(stormi.NewConfigProxy(stormi.NewRedisProxy("127.0.0.1:2131")))
	mp.ConnectByNodeId(33061)
	ct := ConfigTable{}
	ct.Name = "nsqd"
	ct.Addr = "127.0.0.1:" + strconv.Itoa(rand.Intn(10000))
	dtx := mp.NewDTx(id)
	dtx.DB().Create(&ct)
	dtx.Rollback()
}
func server2(id string) {
	mp := stormi.NewMysqlProxy(stormi.NewConfigProxy(stormi.NewRedisProxy("127.0.0.1:2131")))
	mp.ConnectByNodeId(33061)
	ct := ConfigTable{}
	ct.Name = "nsqd"
	ct.Addr = "127.0.0.1:" + strconv.Itoa(rand.Intn(10000))
	dtx := mp.NewDTx(id)
	dtx.DB().Create(&ct)
	dtx.Commit()
}
func server3(id string) {
	mp := stormi.NewMysqlProxy(stormi.NewConfigProxy(stormi.NewRedisProxy("127.0.0.1:2131")))
	mp.ConnectByNodeId(33061)
	ct := ConfigTable{}
	ct.Name = "nsqd"
	ct.Addr = "127.0.0.1:" + strconv.Itoa(rand.Intn(10000))
	dtx := mp.NewDTx(id)
	dtx.DB().Create(&ct)
	dtx.Commit()
}

type ConfigTable struct {
	Name string
	Addr string
}
