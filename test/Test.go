package main

import (
	"time"

	"github.com/stormi-li/stormi"
)

func main() {
	cp := stormi.NewServerProxy("127.0.0.1:2131")
	cp.ConfigProxy().AddConfigSyncNotificationHandler(func(configProxy stormi.ConfigProxy, msg string) {})
	cp.Register("stormiserver", "127.0.0.1:8888", 3, 3*time.Second)
	select {}
}
