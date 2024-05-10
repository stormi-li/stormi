package stormi

import (
	"bufio"
	"fmt"
	"os"
	"os/signal"
	"strings"
	"syscall"
	"time"

	"github.com/google/uuid"
)

const chatchannel = "stormi-chat-channel(do not let you boss know)"
const chatlistenchannel = "stormi-chat-listen-channel(do not let you boss know)"
const noredis = "0.0.0.0:0"

type StormiChat struct {
	rp     *RedisProxy
	chatId string
}

func NewStormiChat(rp *RedisProxy) *StormiChat {
	if rp == nil || !rp.isConnected {
		StormiFmtPrintln(magenta, noredis, "无效redis代理")
		return nil
	}
	sc := StormiChat{}
	sc.chatId = uuid.NewString()
	sc.rp = rp
	return &sc
}

func (sc StormiChat) listenChat() {
	pubsub := sc.rp.GetPubSub(chatlistenchannel)
	go func() {
		for {
			sc.rp.Notify(chatlistenchannel, hi)
			time.Sleep(1 * time.Second)
		}
	}()
	count := 0
	var addrmap = make(map[string]time.Time)
	go func() {
		sc.rp.Subscribe(pubsub, 0, func(msg string) int {
			if msg == hi {
				sc.rp.Notify(chatlistenchannel, sc.chatId)
				return 0
			}
			addrmap[msg] = time.Now()
			return 0
		})
	}()
	go func() {
		for {
			c := 0
			for _, t := range addrmap {
				if time.Since(t) < 2*time.Second {
					c++
				}
			}
			if c != count {
				count = c
				fmt.Println(blue+"当前在线人数:", c)
			}
			time.Sleep(1 * time.Second)
		}
	}()
}

func (sc StormiChat) StartSub() {
	StormiFmtPrintln(green, sc.rp.addrs[0], "stormi chat 消息框")
	sc.listenChat()
	pubsub := sc.rp.GetPubSub(chatchannel)
	addSIGINTHandler(func() {
		fmt.Print(reset)
	})
	sc.rp.Subscribe(pubsub, 0, func(msg string) int {
		if len(msg) <= 37 {
			return 0
		}
		parts := strings.Split(msg, "@")
		if parts[len(parts)-1] == sc.chatId {
			return 0
		}
		msg = msg[:len(msg)-37]
		fmt.Println(cyan + msg)
		return 0
	})

}
func (sc StormiChat) StartPub() {
	msgc := make(chan string, 1)
	sd := make(chan struct{})
	go func() {
		sc.rp.Publish(chatchannel, msgc, sd)
	}()
	reader := bufio.NewReader(os.Stdin)
	StormiFmtPrintln(green, sc.rp.addrs[0], "stormi chat 输入框")
	fmt.Print(green)
	addSIGINTHandler(func() {
		sd <- struct{}{}
		fmt.Print(reset)
	})
	for {
		text, _ := reader.ReadString('\n')
		text = strings.TrimSpace(text)
		msgc <- text + "@" + sc.chatId
	}
}

func init() {
	listenSIGINT()
}

var sigintHandlers []func()

func listenSIGINT() {
	go func() {
		sigs := make(chan os.Signal, 1)
		signal.Notify(sigs, syscall.SIGINT, syscall.SIGTERM)
		<-sigs
		for _, h := range sigintHandlers {
			h()
		}
		os.Exit(1)
	}()
}

func addSIGINTHandler(h func()) {
	sigintHandlers = append(sigintHandlers, h)
}
