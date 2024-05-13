package main

import (
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/nsqio/go-nsq"
	"github.com/stormi-li/stormi"
)

func main() {
	np := stormi.NewNsqdProxy(stormi.NewConfigProxy(stormi.NewRedisProxy("127.0.0.1:2131")))
	np.AddConsumeHandler("stormi-nsqd", "channel1", func(message *nsq.Message) error {
		fmt.Println(string(message.Body))
		return nil
	})
	for {
		np.Publish("stormi-nsqd", []byte(uuid.NewString()))
		time.Sleep(200 * time.Millisecond)
	}
}
