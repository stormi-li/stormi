package stormi

import (
	"fmt"
	"time"
)

const (
	reset   = "\x1b[0m"
	red     = "\x1b[31m"
	green   = "\x1b[32m"
	yellow  = "\x1b[33m"
	blue    = "\x1b[34m"
	magenta = "\x1b[35m"
	cyan    = "\x1b[36m"
)

func StormiPrintln(color string, content string) {
	fmt.Println(color + "[stormi:" + time.Now().Truncate(time.Second).Format("2006-01-02 15:04:05") + "]:" + content + reset)
}

func StormiPrint(color string, content string) {
	fmt.Println(color + "[stormi:" + time.Now().Truncate(time.Second).Format("2006-01-02 15:04:05") + "]:" + content + reset)
}

func StormiFmtPrintln(color string, a ...any) {
	fmt.Println(color+"[stormi:"+time.Now().Truncate(time.Second).Format("2006-01-02 15:04:05")+"]:", a, reset)
}

func StormiFmtPrint(color string, a ...any) {
	fmt.Print(color+"[stormi:"+time.Now().Truncate(time.Second).Format("2006-01-02 15:04:05")+"]:", a, reset)
}
