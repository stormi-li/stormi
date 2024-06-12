package main

import (
	"fmt"
	"sync"
)

func main() {
	// stormi.NodeBuilder.Install()

	// cond := sync.NewCond(new(sync.Mutex))
	// go func() {
	// 	time.Sleep(2000)
	// 	cond.Signal()
	// }()
	// cond.L.Lock()
	// cond.Wait()
	// fmt.Println("hhhh")
	// cond.L.Unlock()
	var wg1 sync.WaitGroup

	// var wg2 sync.WaitGroup
	wg1.Add(1)
	wg1.Done()
	wg1.Wait()
	fmt.Println("done")
}
