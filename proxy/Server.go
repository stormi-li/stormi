package proxy

import (
	"fmt"
	"net"
	"os"
	"os/signal"
	"path/filepath"
	"strconv"
	"syscall"

	"github.com/gin-gonic/gin"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials"
	"google.golang.org/grpc/credentials/insecure"
)

var rpc *grpc.Server

var interceptor1 grpc.UnaryServerInterceptor
var interceptor2 grpc.StreamServerInterceptor

func (serverProxy) SetInterceptor(u grpc.UnaryServerInterceptor, s grpc.StreamServerInterceptor) {
	interceptor1 = u
	interceptor2 = s
}

func (serverProxy) GetStormiRpc() *grpc.Server {
	var creds credentials.TransportCredentials

	certificationPath := currentDir + "/server/rpcserver/interceptor" //-----------------------------

	filepath.Walk(certificationPath, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if !info.IsDir() && filepath.Ext(path) == ".pem" {
			certificationPath = path
		}
		return nil
	})

	privatePath := currentDir + "/server/rpcserver/interceptor/private.key" //-----------------------------

	_, err := os.Stat(certificationPath)

	_, err2 := os.Stat(privatePath)
	if err != nil || err2 != nil {
		creds = insecure.NewCredentials()
	} else {
		fmt.Println("已配置证书pem文件和私钥key文件")
		creds, _ = credentials.NewServerTLSFromFile(certificationPath, privatePath)
	}

	rpc = grpc.NewServer(
		grpc.Creds(creds),
		grpc.UnaryInterceptor(interceptor1),
		grpc.StreamInterceptor(interceptor2))
	return rpc
}

func (serverProxy) StartServer() {
	go func() {
		sigs := make(chan os.Signal, 1)
		signal.Notify(sigs, syscall.SIGINT, syscall.SIGTERM)
		<-sigs
		fmt.Println("进行最后一次配置同步工作")
		configSync()
		fmt.Println("配置同步结束", Config.Stormi.RpcServer.Name, "服务关闭")
		rpc.Stop()
		os.Exit(1)
	}()
	listen, err := net.Listen("tcp", ":"+Config.Stormi.RpcServer.Port)
	if err != nil {
		fmt.Println(err.Error())
	}

	fmt.Println(Config.Stormi.RpcServer.Name, "服务启动,监听端口:", Config.Stormi.RpcServer.Port)

	err = rpc.Serve(listen)
	if err != nil {
		fmt.Println(err.Error())
	}
}

var discovered = false

func (serverProxy) GetCloudConn(serverName string, interceptors grpc.UnaryClientInterceptor, creds credentials.TransportCredentials) *grpc.ClientConn {
	if !discovered {
		discover()
		discovered = true
	}
	addr := getServerAddr(serverName)
	conn, _ := grpc.NewClient(addr,
		grpc.WithTransportCredentials(creds),
		grpc.WithChainUnaryInterceptor(interceptors))
	return conn
}

func (serverProxy) StartRpcServer() {
	var dir = sh("stormi-serverdir.sh")
	res := sh("stormi runrpc " + dir + " " + Config.Stormi.RpcServer.Port)
	if res == "1" {
		fmt.Println(Config.Stormi.RpcServer.Name, "服务启动成功，你可以查看rpcrun.log监控其运行状态")
	} else if res == "-1" {
		fmt.Println(Config.Stormi.RpcServer.Name, "服务启动失败，可能是该服务已经启动了,或者端口已被其它进程占用")
	} else if res == "-2" {
		fmt.Println(Config.Stormi.RpcServer.Name, "服务启动失败，server/rpcserver/registerAndStart/RegisterAndStart.go文件不存在")
	} else {
		fmt.Println(res)
	}
}

func (serverProxy) StartWebServer() {
	var dir = sh("stormi-serverdir.sh")
	res := sh("stormi runweb " + dir + " " + Config.Stormi.WebServer.Port)
	if res == "1" {
		fmt.Println(Config.Stormi.WebServer.Name, "服务启动成功，你可以查看webrun.log监控其运行状态")
	} else if res == "-1" {
		fmt.Println(Config.Stormi.WebServer.Name, "服务启动失败，可能是该服务已经启动了,或者端口已被其它进程占用")
	} else if res == "-2" {
		fmt.Println(Config.Stormi.WebServer.Name, "server/webserver/start/Start.go文件不存在")
	} else {
		fmt.Println(res)
	}
}

var r *gin.Engine

func (serverProxy) GetEngine() *gin.Engine {
	if r == nil {
		r = gin.New()
	}
	return r
}

func (serverProxy) GinRun() {
	go func() {
		sigs := make(chan os.Signal, 1)
		signal.Notify(sigs, syscall.SIGINT, syscall.SIGTERM)
		<-sigs
		fmt.Println("进行最后一次配置同步工作")
		configSync()
		fmt.Println("配置同步结束", Config.Stormi.WebServer.Name, "服务关闭")
		os.Exit(1)
	}()
	r.Run(":" + Config.Stormi.WebServer.Port)
}

func (serverProxy) ShutdownRpcServer() {
	port, _ := strconv.Atoi(Config.Stormi.RpcServer.Port)
	res := Shell.PortProcess(port)
	if res == -1 {
		fmt.Println("rpc服务关闭失败, 可能是该服务并未启动, 或者端口发生变更")
	} else {
		r := sh("kill " + strconv.Itoa(res))
		if r == "" {
			fmt.Println("rpc服务关闭成功")
		} else {
			fmt.Println(r)
		}
	}
}

func (serverProxy) ShutdownWebServer() {
	port, _ := strconv.Atoi(Config.Stormi.WebServer.Port)
	res := Shell.PortProcess(port)
	if res == -1 {
		fmt.Println("web服务关闭失败, 可能是该服务并未启动, 或者端口发生变更")
	} else {
		r := sh("kill " + strconv.Itoa(res))
		if r == "" {
			fmt.Println("web服务关闭成功")
		} else {
			fmt.Println(r)
		}
	}
}

type create struct{}

var Create create

func (create) Router(name string) {
	sha("stormi create-router " + name)
}

func (create) Table(name string) {
	sha("stormi create-table " + name)
}

func (create) Queue(name string) {
	sha("stormi create-queue " + name)
}

func (create) ProtoFile(name string) {
	sha("stormi create-proto " + name)
}

func (create) ProtoImpl(name string) {
	sha("stormi proto-impl " + name)
}

func (create) ProtoProxy(name string) {
	sha("stormi proto-proxy " + name)
}
