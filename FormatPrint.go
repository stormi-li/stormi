package stormi

import (
	"fmt"
	"time"
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

// func StormiPrintln(color string, content string) {
// 	fmt.Println(color + "[stormi:" + time.Now().Truncate(time.Second).Format("2006-01-02 15:04:05") + "]:" + content + reset)
// }

// func StormiPrint(color string, content string) {
// 	fmt.Println(color + "[stormi:" + time.Now().Truncate(time.Second).Format("2006-01-02 15:04:05") + "]:" + content + reset)
// }

func StormiFmtPrintln(color string, addr string, a ...any) {
	if loggerLevel == loggerDebug {
		fmt.Println(color+"[stormi-redis:"+addr+" "+time.Now().Truncate(time.Second).Format("2006-01-02 15:04:05")+"]:", a, reset)
	} else if loggerLevel == loggerInfo && color != white && color != yellow {
		fmt.Println(color+"[stormi-redis:"+addr+" "+time.Now().Truncate(time.Second).Format("2006-01-02 15:04:05")+"]:", a, reset)
	} else if loggerLevel == loggerDebug && color == red {
		fmt.Println(color+"[stormi-redis:"+addr+" "+time.Now().Truncate(time.Second).Format("2006-01-02 15:04:05")+"]:", a, reset)
	} else if loggerLevel == loggerIgnore {
		return
	}
}

// func StormiFmtPrint(color string, addr string, a ...any) {
// 	fmt.Print(color+"[stormi-redis:"+addr+" "+time.Now().Truncate(time.Second).Format("2006-01-02 15:04:05")+"]:", a, reset)
// }
