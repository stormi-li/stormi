package stormi

import (
	"fmt"

	"gorm.io/driver/mysql"
	"gorm.io/gorm"
	"gorm.io/gorm/schema"
)

type MysqlProxy struct {
	cp *ConfigProxy
	c  *Config
	db *gorm.DB
}

func NewMysqlProxy(addr any) MysqlProxy {
	mp := MysqlProxy{}
	mp.cp = NewConfigProxy(addr)
	return mp
}

func (mp *MysqlProxy) Register(nodeId int, addr string, username string, password string, dbname string) {
	c := mp.cp.NewConfig()
	c.Name = "mysql"
	c.Addr = addr
	c.Desc = dbname
	c.NodeId = nodeId
	m := make(map[string]string)
	m["username"] = username
	m["password"] = password
	c.Info = m
	mp.cp.Register(c)
}

func (mp *MysqlProxy) ConnectByNodeId(nodeId int) []Config {
	mmap := mp.cp.ConfigSet["mysql"]
	var cs = []Config{}
	for _, c := range mmap {
		if c.NodeId == nodeId {
			cs = append(cs, *c)
		}
	}
	if len(cs) > 1 {
		StormiFmtPrintln(magenta, mp.cp.rdsAddr, "当前节点存在", len(cs), "个mysql配置节点, 请自行选择配置连接数据库")

	} else if len(cs) == 0 {
		StormiFmtPrintln(magenta, mp.cp.rdsAddr, "配置集无法找到mysql配置")
	} else {
		mp.ConnectByConfig(cs[0])
	}

	return cs
}

func (mp *MysqlProxy) ConnectByConfig(c Config) {
	dsn := fmt.Sprintf("%s:%s@tcp(%s)/%s?charset=utf8mb4&parseTime=True&loc=Local&timeout=%s",
		c.Info["username"],
		c.Info["password"],
		c.Addr,
		c.Desc,
		"5s")
	db, err := gorm.Open(mysql.Open(dsn), &gorm.Config{
		SkipDefaultTransaction: true,
		NamingStrategy: schema.NamingStrategy{
			SingularTable: true,
		},
	})
	mp.db = db
	mp.c = &c
	if err != nil {
		StormiFmtPrintln(magenta, mp.cp.rdsAddr, "该配置连接数据库失败:", c.ToJsonStr())
	} else {
		StormiFmtPrintln(yellow, mp.cp.rdsAddr, "数据库连接成功:", c.ToJsonStr())
	}
}

func (mp *MysqlProxy) DB() *gorm.DB {
	if mp.db == nil {
		StormiFmtPrintln(magenta, mp.cp.rdsAddr, "当前代理未连接到数据库")
		return nil
	}
	return mp.db
}

func (mp *MysqlProxy) Config() *Config {
	return mp.c
}

func (mp *MysqlProxy) ConfigProxyD() *ConfigProxy {
	return mp.cp
}

func (mp *MysqlProxy) RedisProxy() *RedisProxy {
	return mp.cp.rp
}

func (mp *MysqlProxy) NewDTM(dtxid string) DTM {
	dtm := DTM{}
	dtm.db = mp.db.Begin()
	dtm.uuid = dtxid
	return dtm
}
