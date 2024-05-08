package stormi

import (
	"fmt"
	"os"

	"github.com/stormi-li/stormi/dirandfileopt"
	"github.com/stormi-li/stormi/formatprint"
	"github.com/stormi-li/stormi/proxy"
)

func Init() {
	dirandfileopt.CreateDirInCurrentDir("app-redis-cluster")
	dirandfileopt.CreateDirInCurrentDir("app-nsqd-cluster")
	dirandfileopt.CreateDirInCurrentDir("app-redis-node")
	dirandfileopt.CreateDirInCurrentDir("mysql")
	dirandfileopt.CreateDirInCurrentDir("nsqd")
	dirandfileopt.CreateDirInCurrentDir("protocode")
	dirandfileopt.CreateDirInCurrentDir("server-rpc")
	dirandfileopt.CreateDirInCurrentDir("server-web")
	dirandfileopt.CreateDirInCurrentDir("server-proxy")
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
		dirandfileopt.AppendToConfigFile(currentDir+"/app.yaml", "  mysql:\n")
		dirandfileopt.AppendToConfigFile(currentDir+"/app.yaml", "    username: root\n")
		dirandfileopt.AppendToConfigFile(currentDir+"/app.yaml", "    password: 123456\n")
		dirandfileopt.AppendToConfigFile(currentDir+"/app.yaml", "    host: \n")
		dirandfileopt.AppendToConfigFile(currentDir+"/app.yaml", "    port: 3306\n")
		dirandfileopt.AppendToConfigFile(currentDir+"/app.yaml", "    dbname: stormi\n")
		dirandfileopt.AppendToConfigFile(currentDir+"/app.yaml", "    timeout: 5s\n")
		dirandfileopt.AppendToConfigFile(currentDir+"/app.yaml", "    skipdefaulttransaction: true\n")
		dirandfileopt.AppendToConfigFile(currentDir+"/app.yaml", "  rpcserver:\n")
		dirandfileopt.AppendToConfigFile(currentDir+"/app.yaml", "    name: \n")
		dirandfileopt.AppendToConfigFile(currentDir+"/app.yaml", "    weight: 1\n")
		dirandfileopt.AppendToConfigFile(currentDir+"/app.yaml", "    port: \n")
		dirandfileopt.AppendToConfigFile(currentDir+"/app.yaml", "  webserver:\n")
		dirandfileopt.AppendToConfigFile(currentDir+"/app.yaml", "    name: \n")
		dirandfileopt.AppendToConfigFile(currentDir+"/app.yaml", "    port: \n")
		dirandfileopt.AppendToConfigFile(currentDir+"/app.yaml", "  redis:\n")
		dirandfileopt.AppendToConfigFile(currentDir+"/app.yaml", "    node: \n")
		dirandfileopt.AppendToConfigFile(currentDir+"/app.yaml", "    nodes: \n")
		formatprint.FormatPrint(formatprint.Blue, "文件创建成功: "+currentDir+"/app.yaml")
	}

	formatprint.FormatPrint(formatprint.Yellow, "stormi初始化完成, 使用前请配置好ip地址")
}

func Exec(cmd string) {
	proxy.ExecCommand(cmd)
}

func Version() {
	fmt.Println("v1.1.1")
}
