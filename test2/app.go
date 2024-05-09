package main

import (
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/stormi-li/stormi"
)

func main() {
	// stormi.RedisProxy.CreateSingleNode(3331, "single node")
	// // stormi.NsqdProxy.CreateNode(4441, "nsqd node")
	// stormi.Version()
	// rdsClient := redis.NewClusterClient(&redis.ClusterOptions{
	// 	Addrs: []string{"192.168.1.103:2221"},
	// })
	// res := rdsClient.ClusterInfo(context.Background())
	// err := res.Err()
	// fmt.Println(err)
	// res1, err := rdsClient.Set(context.Background(), "stormi", "stormi", 10*time.Hour).Result()
	// fmt.Println(res1, err)
	// stormi.Start("C:/Users/lilili/Desktop/stormiframedev/stormi/test2", "192.168.1.103:2221")
	// // stormi.RedisProxy.RedisClient(0)
	// cfg := stormi.NewConfig()
	// cfg.Name = "fsf"
	// cfg.Addr = "123.13.13.13:44"
	// stormi.RegisterConfig(cfg)
	// res := stormi.PullConfig("fsf")
	// stormi.WriteToConfigFile(res)
	// // res1 := stormi.DecoderConfigFile()
	// configProxy := stormi.NewConfigProxy("192.168.1.103:3331")
	// // configProxy.NotifySync("同步新地址")

	// // test1()
	// // configProxy.AddConfigHandler("mysql", func(cmap map[string]stormi.Config) {
	// // 	for name, c := range cmap {
	// // 		fmt.Println(name, c.Addr)
	// // 	}
	// // })
	// configProxy.Info()
	// c := configProxy.NewConfig()
	// c.Name = "mysql"
	// c.Addr = "123.13.31.31:43"
	// configProxy.Register(c)
	// configProxy.SyncConfig()
	// select {}
	// sp := stormi.NewServerProxy("192.168.1.103:2221")
	// // sp.Register("server", "123.13.31.13:434", 3, 3*time.Second)
	// // time.Sleep(2 * time.Second)
	// sp.Discover("server", 100*time.Millisecond, func(addr string) error {
	// 	fmt.Println(addr)
	// 	return nil
	// })
	// select {}
	// rp := stormi.NewRedisProxy("192.168.1.103:2221")
	// go func() {
	// 	l := rp.NewLock("lock")
	// 	for {
	// 		l.Lock()
	// 		fmt.Println("1")
	// 		time.Sleep(2 * time.Second)
	// 		l.UnLock()
	// 		time.Sleep(100 * time.Millisecond)

	// 	}

	// }()
	// go func() {
	// 	l := rp.NewLock("lock")
	// 	for {
	// 		l.Lock()
	// 		fmt.Println("2")
	// 		time.Sleep(2 * time.Second)
	// 		l.UnLock()
	// 		time.Sleep(100 * time.Millisecond)

	// 	}

	// }()
	// select {}

	// tp := stormi.NewTransactionProxy("192.168.1.103:2221")
	// ids := tp.NewDTxIds(3)
	// func1(ids[0])
	// func2(ids[1])
	// func3(ids[2])
	// tp.DCommit(ids, func(statement [][2]string) {
	// 	fmt.Println("事务失败")
	// 	fmt.Println(statement)
	// })
	// select {}
	// var b = []string{}
	// for _, a := range b {
	// 	fmt.Println(a)
	// }
	// fmt.Println("hh")

	// np := stormi.NewNsqdProxy("192.168.1.103:2221")
	// // np.Register("192.168.1.103:9092")
	// // go func() {

	// // }()
	// np.AddConsumeHandler("hhhh", "abc", func(message *nsq.Message) error {
	// 	fmt.Println(string(message.Body))
	// 	return nil
	// })
	// np.AddConsumeHandler("hhhh", "abc", func(message *nsq.Message) error {
	// 	fmt.Println(string(message.Body), 2)
	// 	return nil
	// })
	// go func() {
	// 	time.Sleep(3 * time.Second)
	// 	np.Stop()
	// }()
	// for {
	// 	np.Publish("hhhh", []byte(uuid.NewString()))
	// 	time.Sleep(1000 * time.Millisecond)
	// }
	// cp := stormi.NewConfigProxy("192.168.1.103:2221")
	// cp.RemoveRegister("server")
	// cmap := cp.ConfigSet["server"]
	// for _, c := range cmap {

	// 	cp.Remove(*c)

	// }
	// np.Register("192.168.1.103:4150")

	// t := New()
	// t.Print()
	// stormi.NodeBuilder.CreateRedisNode(1112, stormi.NodeType.RedisCluster, "C:\\Users\\lilili\\Desktop\\nsqd\\2", "192.168.1.103")
	// stormi.NodeBuilder.AddNodeToRedisCluster("192.168.1.103:1112", "192.168.1.103:2221", stormi.NodeType.RedisSlave)
	// stormi.NodeBuilder.CreateRedisCluster(8881, 8882, 8883, 8884, 8885, 8886, "192.168.1.103", "C:\\Users\\lilili\\Desktop\\rediscluster")
	stormi.StartHardWord(stormi.NewRedisProxy("192.168.1.103:2221"))
}

type Test struct {
	m string
}

func New() *Test {
	t1 := Test{}
	t1.change()
	return &t1
}
func (t *Test) change() {
	go func() {
		t.m = "hh"
	}()
}

func (t *Test) Print() {
	for {
		fmt.Println(t.m)
		time.Sleep(1 * time.Second)
	}
}

func func1(id string) {
	mp := stormi.NewMysqlProxy("192.168.1.103:2221")
	mp.ConnectByNodeId(1234)
	dtx := mp.NewDTx(id)
	dtx.DB().Create(&User{Name: uuid.NewString() + "1", Age: 1})
	dtx.Commit()
}

func func2(id string) {
	mp := stormi.NewMysqlProxy("192.168.1.103:2221")
	mp.ConnectByNodeId(1234)
	dtx := mp.NewDTx(id)
	dtx.DB().Create(&User{Name: uuid.NewString() + "2", Age: 2})
	dtx.Commit()
}

func func3(id string) {
	mp := stormi.NewMysqlProxy("192.168.1.103:2221")
	mp.ConnectByNodeId(1234)
	dtx := mp.NewDTx(id)
	dtx.DB().Create(&User{Name: uuid.NewString() + "3", Age: 3})
	dtx.Rollback()
	dtx.Commit()

}

type User struct {
	ID   uint
	Name string
	Age  uint
}

func test1() {
	configProxy := stormi.NewConfigProxy("192.168.1.103:2221")
	configProxy.SetConfigSyncNotficationHandler(func(configProxy stormi.ConfigProxy, msg string) {
		fmt.Println("不同步")
	})
}
