package formatprint

import (
	"fmt"
	"time"
)

const (
	Reset   = "\x1b[0m"
	Red     = "\x1b[31m"
	Green   = "\x1b[32m"
	Yellow  = "\x1b[33m"
	Blue    = "\x1b[34m"
	Magenta = "\x1b[35m"
	Cyan    = "\x1b[36m"
)

func FormatPrint(color string, content string) {
	fmt.Println(color + "[stormi:" + time.Now().Truncate(time.Second).Format("2006-01-02 15:04:05") + "]:" + content + Reset)
}
