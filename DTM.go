package stormi

import (
	"fmt"

	"gorm.io/gorm"
)

func Test() {}

type DTM struct {
	db         *gorm.DB
	rollbacked bool
	uuid       string
}

func (dtm *DTM) NewDTx(dtxid string) *DTM {
	dtm.uuid = dtxid
	return dtm
}

func (dtm *DTM) DB() *gorm.DB {
	return dtm.db
}

func (dtm *DTM) Rollback() {
	if dtm.rollbacked {
		return
	} else {
		dtm.rollbacked = true

	}
}
func (dtm *DTM) Commit() {
	if dtm.rollbacked {
		return
	} else {
		fmt.Println("commit")
	}
}
