package stormi

import (
	"fmt"
)

func GitHubUrl() {
	fmt.Println(green+"源代码地址: https://github.com/stormi-li/stormi", reset)
	fmt.Println(green+"资源地址: https://github.com/stormi-li/stormi-bin-resource", reset)
}

func SetLoggerLevel(level int) {
	loggerLevel = level
}
