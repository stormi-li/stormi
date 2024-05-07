package stormi

import (
	"os"

	"github.com/stormi-li/stormi/dirandfileopt"
	"github.com/stormi-li/stormi/formatprint"
	"github.com/stormi-li/stormi/proxy"
)

var RedisProxy = proxy.RedisProxy

func Init() {
	dirandfileopt.CreateDirInCurrentDir("app-redis-cluster")
	dirandfileopt.CreateDirInCurrentDir("app-nsqd-cluster")
	dirandfileopt.CreateDirInCurrentDir("app-redis-node")
	currentDir, _ := os.Getwd()
	if dirandfileopt.ExistFile(currentDir + "/app.config") {
		formatprint.FormatPrint(formatprint.Magenta, "文件已存在: "+currentDir+"/app.config")
	} else {
		dirandfileopt.AppendToConfigFile(currentDir+"/app.config", "")
		formatprint.FormatPrint(formatprint.Blue, "文件创建成功: "+currentDir+"/app.config")
	}
	if dirandfileopt.ExistFile(currentDir + "/app.yaml") {
		formatprint.FormatPrint(formatprint.Magenta, "文件已存在: "+currentDir+"/app.yaml")
	} else {
		dirandfileopt.AppendToConfigFile(currentDir+"/app.yaml", "stormi:\n")
		dirandfileopt.AppendToConfigFile(currentDir+"/app.yaml", "  ip: \n")
		dirandfileopt.AppendToConfigFile(currentDir+"/app.yaml", "  redis:\n")
		dirandfileopt.AppendToConfigFile(currentDir+"/app.yaml", "    nodes: \n")
		formatprint.FormatPrint(formatprint.Blue, "文件创建成功: "+currentDir+"/app.yaml")
	}
	formatprint.FormatPrint(formatprint.Yellow, "stormi初始化完成")
}

func Exec(cmd string) {
	proxy.ExecCommand(cmd)
}
