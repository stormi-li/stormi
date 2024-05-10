package main

import (
	"github.com/stormi-li/stormi"
)

func main() {
	cp := stormi.NewServerProxy("127.0.0.1:2131")
	cp.ConfigProxy().ConfigPersistence()
}
