package main

import (
	"fmt"
	"sync"

	"github.com/stormi-li/stormi"
)

func main() {
	syp := stormi.NewSyncProxy(stormi.NewRedisProxy("127.0.0.1:213"))
	cond := syp.NewCond("cond")
	var wg sync.WaitGroup
	for {
		wg.Add(3)
		go func() {
			cond.Wait()
			fmt.Println("wait1")
			wg.Done()
		}()
		go func() {
			cond.Wait()
			fmt.Println("wait2")
			wg.Done()
		}()
		go func() {
			cond.Wait()
			fmt.Println("wait3")
			wg.Done()
		}()
		wg.Wait()
	}
}
