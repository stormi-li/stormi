package stormi

import (
	"github.com/google/uuid"
)

type CooperationProxy struct {
	cooperationName string
	cp              *ConfigProxy
	uuid            string
}

var coprotocol = "coprotocol"

func NewCooperationProxy(cp *ConfigProxy, name string) *CooperationProxy {
	if !isFirstCharUpperCaseLetter(name) {
		StormiFmtPrintln(magenta, noredis, "错误命名:", name)
		return nil
	}
	cop := CooperationProxy{}
	cop.cp = cp
	cop.cooperationName = name
	cop.pullCoprotocol()
	return &cop
}

func (cop *CooperationProxy) Register() {
	if cop.cp.IsRegistered(cop.cooperationName + coprotocol) {
		StormiFmtPrintln(magenta, cop.cp.rdsAddr, "名字:"+cop.cooperationName+"已被注册")
		return
	}
	c := cop.cp.NewConfig()
	c.Name = cop.cooperationName
	cop.cp.Register(c)
	c.Name = cop.cooperationName + coprotocol
	cop.cp.Register(c)
}

func (cop *CooperationProxy) CreateCoprotocol() {
	FileOpt.createCoprotocol(cop.cooperationName)
}

func (cop *CooperationProxy) PushCoprotocol() {
	procode := FileOpt.decodeCoprotocol(cop.cooperationName)
	c := cop.cp.NewConfig()
	c.Name = cop.cooperationName + coprotocol
	cmap := cop.cp.Pull(c.Name)
	if len(cmap) == 0 {
		c.Addr = uuid.NewString()
		c.Desc = procode
		cop.cp.Register(c)
	} else {
		for _, cc := range cmap {
			c = cc
			break
		}
		c.Desc = procode
		cop.cp.Update(c)
	}
	cop.uuid = c.Addr
}

func (cop *CooperationProxy) pullCoprotocol() {
	cmap := cop.cp.Pull(cop.cooperationName + coprotocol)
	if len(cmap) == 0 {
		StormiFmtPrintln(magenta, cop.cp.rdsAddr, "协议不存在")
	} else {
		var procode string
		for _, c := range cmap {
			procode = c.Desc
			cop.uuid = c.Addr
			break
		}
		path := FileOpt.GoModDir() + "/" + coprotocol + "/" + cop.cooperationName
		FileOpt.Mkdir(path)
		filename := path + "/" + cop.cooperationName + ".go"
		FileOpt.TruncateFile(filename)
		FileOpt.AppendToFile(filename, procode)
	}
}

func (cop *CooperationProxy) NewHandler() {
	cop.pullCoprotocol()

}
func (cop *CooperationProxy) NewCaller() {
	cop.pullCoprotocol()
}
