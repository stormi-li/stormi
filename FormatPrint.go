package stormi

import (
	"fmt"
	"time"
)

const (
	reset     = "\x1b[0m"
	bold      = "\x1b[1m"
	dim       = "\x1b[2m"
	italic    = "\x1b[3m"
	underline = "\x1b[4m"
	blink     = "\x1b[5m"
	invert    = "\x1b[7m"
	strike    = "\x1b[9m"

	black   = "\x1b[30m"
	red     = "\x1b[31m"
	green   = "\x1b[32m"
	yellow  = "\x1b[33m"
	blue    = "\x1b[34m"
	magenta = "\x1b[35m"
	cyan    = "\x1b[36m"
	white   = "\x1b[37m"

	// bgBlack   = "\x1b[40m"
	// bgRed     = "\x1b[41m"
	// bgGreen   = "\x1b[42m"
	// bgYellow  = "\x1b[43m"
	// bgBlue    = "\x1b[44m"
	// bgMagenta = "\x1b[45m"
	// bgCyan    = "\x1b[46m"
	// bgWhite   = "\x1b[47m"
)

var loggerLevel = 0

var LoggerLevel = loggerlevel{}

type loggerlevel struct {
	LoggerDebug  int
	LoggerInfo   int
	LoggerError  int
	LoggerIgnore int
}

func init() {
	LoggerLevel.LoggerDebug = loggerDebug
	LoggerLevel.LoggerInfo = loggerInfo
	LoggerLevel.LoggerError = loggerError
	LoggerLevel.LoggerIgnore = loggerIgnore
}

const loggerIgnore = -1
const loggerDebug = 0
const loggerInfo = 1
const loggerError = 2

func StormiPrintln(color string, content string) {
	fmt.Println(color + "[stormi:" + time.Now().Truncate(time.Second).Format("2006-01-02 15:04:05") + "]:" + content + reset)
}

func StormiPrint(color string, content string) {
	fmt.Println(color + "[stormi:" + time.Now().Truncate(time.Second).Format("2006-01-02 15:04:05") + "]:" + content + reset)
}

func StormiFmtPrintln(color string, addr string, a ...any) {
	if loggerLevel == loggerDebug {
		fmt.Println(color+"[stormi-redis:"+addr+" "+time.Now().Truncate(time.Second).Format("2006-01-02 15:04:05")+"]:", a, reset)
	} else if loggerLevel == loggerInfo && color != white && color != green {
		fmt.Println(color+"[stormi-redis:"+addr+" "+time.Now().Truncate(time.Second).Format("2006-01-02 15:04:05")+"]:", a, reset)
	} else if loggerLevel == loggerDebug && color == red {
		fmt.Println(color+"[stormi-redis:"+addr+" "+time.Now().Truncate(time.Second).Format("2006-01-02 15:04:05")+"]:", a, reset)
	}
}

func StormiFmtPrint(color string, addr string, a ...any) {
	fmt.Print(color+"[stormi-redis:"+addr+" "+time.Now().Truncate(time.Second).Format("2006-01-02 15:04:05")+"]:", a, reset)
}
