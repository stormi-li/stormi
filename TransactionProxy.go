package stormi

import (
	"strconv"
	"strings"
	"time"

	"github.com/google/uuid"
)

const (
	rollbackwaiting = "rollbackwaiting"
	commitwaiting   = "commitwaiting"
	rollback        = "rollback"
	commit          = "commit"
	noresponse      = "noresponse"
	report          = "report"
	finished        = "finished"
	hi              = "hi"
)

type TransactionProxy struct {
	rp *RedisProxy
}

type dTxStatment struct {
	Committed       string
	CommitWaiting   string
	Rollbacked      string
	RollbackWaiting string
	NoResponse      string
}

var DTxStatment dTxStatment

func init() {
	DTxStatment = dTxStatment{}
	DTxStatment.Committed = "Committed"
	DTxStatment.CommitWaiting = "CommitWaiting"
	DTxStatment.Rollbacked = "Rollbacked"
	DTxStatment.RollbackWaiting = "RollbackWaiting"
	DTxStatment.NoResponse = "NoResponse"

}

func NewTransactionProxy(addr any) *TransactionProxy {
	tp := TransactionProxy{}
	rp, ok := addr.(*RedisProxy)
	if ok {
		tp.rp = rp
	} else {
		tp.rp = NewRedisProxy(addr)
	}
	return &tp
}

func (tp TransactionProxy) RedisProxy(num int) *RedisProxy {
	if tp.rp != nil {
		return tp.rp
	} else {
		StormiFmtPrintln(magenta, noredis, "未初始化redis代理")
		return nil
	}
}

func (tp TransactionProxy) NewDTxIds(num int) []string {
	ids := []string{}
	uuid := uuid.NewString()
	for i := 0; i < num; i++ {
		ids = append(ids, uuid+"@"+strconv.Itoa(i))
	}
	return ids
}

func (tp TransactionProxy) DCommit(dtxids []string, handler func(statement [][2]string)) {
	num := len(dtxids)
	if num == 0 {
		StormiFmtPrintln(magenta, tp.rp.addrs[0], "无效事务ids:", dtxids)
		return
	}
	uuid := ""
	for _, id := range dtxids {
		parts := strings.Split(id, "@")
		if len(parts) != 2 {
			StormiFmtPrintln(magenta, tp.rp.addrs[0], "无效事务id:", id)
			return
		} else {
			if uuid != "" && parts[0] != uuid {
				StormiFmtPrintln(magenta, tp.rp.addrs[0], "事务id不一致:", id, uuid)
				return
			}
		}
		if uuid == "" {
			uuid = parts[0]
		}
	}
	statement := make([][2]string, num)
	for i := 0; i < num; i++ {
		statement[i][0] = DTxStatment.NoResponse
		statement[i][1] = DTxStatment.NoResponse
	}
	cmtcount := 0
	fnscount := 0
	pubsub := tp.rp.GetPubSub(uuid)
	iscommit := false
	allcommit := true
	tp.rp.Notify(uuid, report)
	go func() {
		res := tp.rp.Subscribe(pubsub, 3*time.Second, func(msg string) int {
			parts := strings.Split(msg, "@")
			if len(parts) == 2 {
				index, _ := strconv.Atoi(parts[1])
				if parts[0] == rollbackwaiting {
					statement[index][0] = DTxStatment.RollbackWaiting
					tp.rp.Notify(uuid, rollback)
				}
				if parts[0] == commitwaiting {
					cmtcount++
					statement[index][0] = DTxStatment.CommitWaiting
					if cmtcount == num {
						tp.rp.Notify(uuid, commit)
						iscommit = true
					}
				}
			}
			if len(parts) == 3 {
				index, _ := strconv.Atoi(parts[1])
				if parts[2] == rollback {
					statement[index][1] = DTxStatment.Rollbacked
					fnscount++
					if iscommit {
						allcommit = false
					}
				}
				if parts[2] == commit {
					statement[index][1] = DTxStatment.Committed
					fnscount++
				}
				if fnscount == num {
					if iscommit && allcommit {
						return 1
					} else {
						return -1
					}
				}
			}
			return 0
		})
		if res != 1 {
			handler(statement)
		}
	}()
}
