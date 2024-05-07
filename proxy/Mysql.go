package stormi

import (
	"context"
	"errors"
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/google/uuid"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
	"gorm.io/gorm/schema"
)

var db *gorm.DB

func mysqlInit() {
	dsn := fmt.Sprintf("%s:%s@tcp(%s:%s)/%s?charset=utf8mb4&parseTime=True&loc=Local&timeout=%s",
		Config.Stormi.Mysql.Username,
		Config.Stormi.Mysql.Password,
		Config.Stormi.Mysql.Host,
		Config.Stormi.Mysql.Port,
		Config.Stormi.Mysql.Dbname,
		Config.Stormi.Mysql.Timeout)
	_, err := gorm.Open(mysql.Open(dsn), &gorm.Config{
		SkipDefaultTransaction: Config.Stormi.Mysql.Skipdefaulttransaction,
		NamingStrategy: schema.NamingStrategy{
			SingularTable: true,
		},
	})
	if err != nil {
		fmt.Println(red, "[error]: failed to connect database, err: ", err, reset)
	}
}

type mysqlManager struct {
}

var MysqlProxy mysqlManager

func (mysqlManager) DB() *gorm.DB {
	return db
}

type mysqlTransaction struct {
	Tranction string
}
type txManager struct {
	DB         *gorm.DB
	Rollbacked bool
	DTXid      string
	Err        error
}

func (mysqlManager) DBConfig(mysqlCfg *gorm.Session) {
	db = db.Session(mysqlCfg)
}

func (mysqlManager) NewTxCtx(TxId string) context.Context {
	ctx := context.Background()
	tm := txManager{
		DB:         db.Begin(),
		Rollbacked: false,
		DTXid:      TxId,
	}
	ctx = context.WithValue(ctx, mysqlTransaction{}, &tm)
	return ctx
}

func (mysqlManager) ParseTxContext(txctx context.Context) (*gorm.DB, bool, string, error) {
	tm := txctx.Value(mysqlTransaction{})
	txManager, ok := tm.(*txManager)
	if ok {
		return txManager.DB, txManager.Rollbacked, txManager.DTXid, txManager.Err
	}
	return nil, false, "", errors.New("mysql transaction manager not found")
}

func (mysqlManager) SetRollback(txctx context.Context, err error) {
	tm := txctx.Value(mysqlTransaction{})
	txManager := tm.(*txManager)
	txManager.Rollbacked = true
	txManager.Err = err
}

func (mysqlManager) NewDTxId(num int) []string {
	ids := []string{}
	uuid := uuid.NewString()
	ids = append(ids, uuid+"@"+strconv.Itoa(num))
	for i := 1; i <= num; i++ {
		ids = append(ids, uuid+"@"+strconv.Itoa(i))
	}
	return ids
}

func (mysqlManager) TxCommit(txctx context.Context, handler func(code int, err error)) {
	txdb, rollbacked, id, err := MysqlProxy.ParseTxContext(txctx)

	if id == "" {
		if err != nil {
			handler(0, err)
		} else {
			txdb.Commit()
		}
	} else {
		msgchan := make(chan string, 1)
		shutdown := make(chan struct{})
		parts := strings.Split(id, "@")
		sub := rdsClusterClient.Subscribe(context.Background(), id)
		go func() {
			publish(parts[0], msgchan, shutdown)
		}()

		go func() {
			res := subscribe(sub.Channel(), 3*time.Second, func(msg string) int {
				if msg == "sync-1" {
					msgchan <- "sync-1-received@" + parts[1]
				}
				if msg == "sync-2" {
					if rollbacked {
						msgchan <- "sync-2-received@" + parts[1] + "@rollbacked"
					} else {
						msgchan <- "sync-2-received@" + parts[1] + "@waiting"
					}
				}
				if msg == "rollback" {
					msgchan <- "status@" + parts[1] + "@finished"
					return -1
				}
				if msg == "commit" {
					msgchan <- "status@" + parts[1] + "@finished"
					return 1
				}
				return 0
			})

			shutdown <- struct{}{}
			if res != 1 {
				txdb.Rollback()
				handler(res, err)
			} else {
				txdb.Commit()
			}
		}()
	}
}

func (mysqlManager) DTxCommit(id string, handler func(code int, txStatus []string, dTxStatus []string)) {
	var msgchanList []chan string
	var shutdownList []chan struct{}
	parts := strings.Split(id, "@")
	sentryId := parts[0]
	num, _ := strconv.Atoi(parts[1])
	sub := rdsClusterClient.Subscribe(context.Background(), sentryId)
	for i := 0; i < num; i++ {
		msgchanList = append(msgchanList, make(chan string, 1))
		msgchanList[i] <- "sync-1"
		shutdownList = append(shutdownList, make(chan struct{}))
		go func(i int) {
			publish(sentryId+"@"+strconv.Itoa(i+1), msgchanList[i], shutdownList[i])
		}(i)
	}
	var status []string
	var status2 []string
	count := 0
	rollback := false
	for i := 0; i < num; i++ {
		status = append(status, "")
		status2 = append(status2, "")
	}
	go func() {
		res := subscribe(sub.Channel(), 3*time.Second, func(msg string) int {
			parts := strings.Split(msg, "@")
			if parts[0] == "sync-1-received" {
				index, _ := strconv.Atoi(parts[1])
				msgchanList[index-1] <- "sync-2"
			}
			if parts[0] == "sync-2-received" {
				index, _ := strconv.Atoi(parts[1])
				status[index-1] = parts[2]
				if parts[2] == "rollbacked" {
					rollback = true
				}
				count++
				if count == num {
					if rollback {
						for i := 0; i < num; i++ {
							msgchanList[i] <- "rollback"
						}
					} else {
						for i := 0; i < num; i++ {
							msgchanList[i] <- "commit"
						}
					}
					count = 0
				}
			}
			if parts[0] == "status" {
				index, _ := strconv.Atoi(parts[1])
				status2[index-1] = parts[2]
				count++
				if count == num {
					if rollback {
						return -1
					} else {
						return 1
					}
				}
			}
			return 0
		})

		for i := 0; i < num; i++ {
			shutdownList[i] <- struct{}{}
		}

		if res != 1 {
			handler(res, status, status2)
		}
	}()
}
