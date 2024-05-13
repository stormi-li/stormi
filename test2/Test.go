package main

import (
	"bufio"
	"os"
	"strings"

	"github.com/stormi-li/stormi"
)

func main() {
	rp := stormi.NewRedisProxy("127.0.0.1:2131")
	reader := bufio.NewReader(os.Stdin)
	for {
		text, _ := reader.ReadString('\n')
		text = strings.TrimSpace(text)
		rp.Notify("stormi-chat", text)
	}
}
