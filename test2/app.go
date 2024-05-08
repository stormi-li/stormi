package main

import "github.com/stormi-li/stormi"

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
	// res1 := stormi.DecoderConfigFile()
	configProxy := stormi.NewConfigProxy("192.168.1.103:2221")
	configProxy.Info()

}
