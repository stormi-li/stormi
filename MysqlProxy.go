package stormi

import (
	"fmt"
	"strconv"
	"strings"
	"time"

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

func (mp *MysqlProxy) ConfigProxy() *ConfigProxy {
	return mp.cp
}

func (mp *MysqlProxy) RedisProxy() *RedisProxy {
	return mp.cp.rp
}

type DTx struct {
	db         *gorm.DB
	rollbacked bool
	committed  bool
	uuid       string
	index      int
	rp         *RedisProxy
}

func (mp MysqlProxy) NewDTx(dtxid string) *DTx {
	dtx := DTx{}
	dtx.uuid = dtxid
	parts := strings.Split(dtxid, "@")
	if len(parts) != 2 {
		StormiFmtPrintln(magenta, "无效事务id:", dtxid)
		return nil
	}
	dtx.uuid = parts[0]
	index, err := strconv.Atoi(parts[1])
	if err != nil {
		StormiFmtPrintln(magenta, "无效事务id:", dtxid)
		return nil
	}
	dtx.index = index
	dtx.db = mp.db.Begin()
	dtx.rp = mp.cp.rp
	return &dtx
}

func (dtx *DTx) DB() *gorm.DB {
	return dtx.db
}

func (dtx *DTx) Rollback() {
	if dtx.rollbacked || dtx.committed {
		return
	} else {
		dtx.rollbacked = true
		dtx.dtxhandle()
	}
}
func (dtx *DTx) Commit() {
	if dtx.rollbacked || dtx.committed {
		return
	} else {
		dtx.committed = true
		dtx.dtxhandle()
	}
}

func (dtx *DTx) dtxhandle() {
	pubsub := dtx.rp.GetPubSub(dtx.uuid)
	go func() {
		index := strconv.Itoa(dtx.index)
		hiindex := hi + "@" + index
		status := ""
		if dtx.rollbacked {
			status = rollbackwaiting + "@" + index
		} else {
			status = commitwaiting + "@" + index
		}
		finishm := finished + "@" + index
		dtx.rp.Notify(dtx.uuid, hi)
		resp := dtx.rp.Subscribe(pubsub, 10*time.Second, func(msg string) int {
			if msg == rollback {
				dtx.rp.Notify(dtx.uuid, finishm+"@"+rollback)
				return 2
			}
			if msg == commit {
				dtx.rp.Notify(dtx.uuid, finishm+"@"+commit)
				return 1
			}
			if msg == hi {
				dtx.rp.Notify(dtx.uuid, hiindex)
			}
			if msg == report {
				dtx.rp.Notify(dtx.uuid, status)
			}
			return 0
		})
		if resp == 1 {
			dtx.db.Commit()
		} else {
			dtx.db.Rollback()
		}
	}()
}
