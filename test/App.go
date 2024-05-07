package main

import "github.com/stormi-li/stormi"

func main() {
	// stormi.RedisProxy.C,"reateClusterNode(2334, "fsdfs")
	// output, err := exec.Command("cmd", "/C", "stormi version").CombinedOutput()
	// if err != nil {
	// 	fmt.Println("执行脚本时出错：", err)
	// 	return
	// }

	// // 输出脚本执行的结果
	// fmt.Println("脚本输出：", string(output))
	// stormi.Exec("stormi version")
	// stormi.RedisProxy.CreateClusterNode(4343, "desc")
	// stormi.RedisProxy.CreateCluster(7771, 7772, 7773, 7774, 7775, 7776)
	stormi.RedisProxy.StartSingleNode(2334)
}
