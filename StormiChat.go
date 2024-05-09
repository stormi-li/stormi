package stormi

import (
	"bufio"
	"fmt"
	"os"
	"strings"

	"github.com/google/uuid"
)

const hardwork = "stormi-hard-work-channel(do not let you boss know)"

func StartHardWord(rp *RedisProxy) {
	id := uuid.NewString()
	pubsub := rp.GetPubSub(hardwork)
	go func() {
		rp.Subscribe(pubsub, 0, func(msg string) int {
			if len(msg) <= 37 {
				return 0
			}
			parts := strings.Split(msg, "@")
			if parts[len(parts)-1] == id {
				return 0
			}
			msg = msg[:len(msg)-37]
			fmt.Println(cyan+msg, green)
			return 0
		})
	}()
	msgc := make(chan string, 1)
	sd := make(chan struct{})
	go func() {
		rp.Publish(hardwork, msgc, sd)
	}()
	reader := bufio.NewReader(os.Stdin)
	StormiFmtPrintln(green, rp.addrs[0], "请开始努力工作吧, exit退出努力工作")
	fmt.Print(green)
	for {
		text, _ := reader.ReadString('\n')
		text = strings.TrimSpace(text)
		if text == "exit" {
			sd <- struct{}{}
			break
		}
		msgc <- text + "@" + id
	}
	fmt.Print(reset)
}
