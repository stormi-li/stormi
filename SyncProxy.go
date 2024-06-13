package stormi

import (
	"context"
	"os"
	"strconv"
	"time"

	"github.com/go-redis/redis/v8"
)

//分布式读写锁，分布式WaitGroup,分布式条件变量

type SyncProxy struct {
	rp *RedisProxy
}

func NewSyncProxy(rp *RedisProxy) *SyncProxy {
	if rp.isCluster {
		StormiFmtPrintln(red, rp.addrs[0], "目前框架并不能很好的解决redis集群中可能出现的数据一致性问题, sync代理暂不支持redist集群模式")
		return nil
	}
	sp := SyncProxy{}
	sp.rp = rp
	return &sp
}

const rwLockPrefix = "stormi-sync-rwlock-"

type DRWLock struct {
	lockName string
	lockInfo string
	rp       *RedisProxy
	rLock    DLock
	rLocked  bool
	wLock    DLock
}

func (sp SyncProxy) NewRWLock(lockName string) *DRWLock {
	dwrl := DRWLock{}
	dwrl.lockName = rwLockPrefix + lockName
	dwrl.lockInfo = rwLockPrefix + lockName + "-rLock:" + lockInfo()
	dwrl.rp = sp.rp
	dwrl.rLock = *sp.rp.NewLock(dwrl.lockInfo)
	dwrl.wLock = *sp.rp.NewLock(rwLockPrefix + lockName + "-wLock")
	return &dwrl
}

func (dwrl *DRWLock) RLock() {
	if dwrl.rLocked {
		return
	}
	dwrl.wLock.Lock()
	dwrl.rLock.Lock()
	dwrl.wLock.UnLock()
	dwrl.rLocked = true
}

func (dwrl *DRWLock) RUnlock() {
	dwrl.rLock.UnLock()
	if !NamespaceHasKeys(dwrl.rp.rdsClient, context.Background(), dwrl.lockName+"-rLock") {
		dwrl.rp.Notify(dwrl.lockName, "Unlock")
	}
}

func (dwrl *DRWLock) WLock() {
	dwrl.wLock.Lock()
	for {
		if !NamespaceHasKeys(dwrl.rp.rdsClient, context.Background(), dwrl.lockName+"-rLock") {
			break
		}
		res := dwrl.rp.Wait(dwrl.lockName, 3*time.Second)
		if res == "Unlock" {
			break
		}
	}
}

func (dwrl *DRWLock) WUnlock() {
	dwrl.wLock.UnLock()
}

func lockInfo() string {
	info := getIp() + ":" + strconv.Itoa(os.Getpid())
	return info
}

type WaitGroup struct {
	groupName string
	lock      *DLock
	rp        *RedisProxy
}

const waitGroupPrefix = "stormi-sync-waitgroup-"

func (sp *SyncProxy) NewWaitGroup(groupName string) *WaitGroup {
	wg := WaitGroup{}
	wg.groupName = waitGroupPrefix + groupName
	wg.lock = sp.rp.NewLock(wg.groupName + "-lock")
	wg.rp = sp.rp
	return &wg
}

func (wg *WaitGroup) Add(delta int) {
	wg.lock.Lock()
	defer wg.lock.UnLock()
	ctx := context.Background()
	key := wg.groupName + "-count"
	val, err := wg.rp.rdsClient.Get(ctx, key).Result()
	count := 0
	if err == redis.Nil {
		count = delta
	} else {
		vali, _ := strconv.Atoi(val)
		count = vali + delta
	}
	if count == 0 {
		wg.rp.rdsClient.Del(ctx, key)
		wg.rp.Notify(key, "done")
	} else if count < 0 {
		panic("sync: negative WaitGroup counter")
	} else if count > 0 {
		wg.rp.rdsClient.Set(ctx, key, count, 0)
	}
}

func (wg *WaitGroup) Done() {
	wg.Add(-1)
}

func (wg *WaitGroup) Wait() {
	key := wg.groupName + "-count"
	ctx := context.Background()
	for {
		_, err := wg.rp.rdsClient.Get(ctx, key).Result()
		if err == redis.Nil {
			break
		}
		res := wg.rp.Wait(key, 3*time.Second)
		if res == "done" {
			break
		}
	}
}

type Condition struct {
	conditionName string
	lock          *DLock
	rp            *RedisProxy
}

const conditionPrefix = "stormi-sync-Condition-"

func (sp *SyncProxy) NewCond(conditionName string) *Condition {
	cd := Condition{}
	cd.conditionName = conditionPrefix + conditionName
	cd.lock = sp.rp.NewLock(conditionPrefix + "-lock")
	cd.rp = sp.rp
	return &cd
}

func (cond *Condition) Wait() {
	key := cond.conditionName + "-count"
	ctx := context.Background()
	for {
		cond.lock.Lock()
		val, err := cond.rp.RedisClient().Get(ctx, key).Result()

		if err != redis.Nil {
			count, _ := strconv.Atoi(val)
			if count > 0 {
				count--
				if count == 0 {
					cond.rp.RedisClient().Del(ctx, key)
					cond.lock.UnLock()
					break
				}
				cond.rp.RedisClient().Set(ctx, key, count, 0)
				cond.lock.UnLock()
				break
			}
		}
		cond.lock.UnLock()
		res := cond.rp.Wait(key, 3*time.Second)
		if res == "broadcast" {
			break
		}
	}
}

func (cond *Condition) Singal() {
	key := cond.conditionName + "-count"
	ctx := context.Background()

	cond.lock.Lock()
	val, err := cond.rp.RedisClient().Get(ctx, key).Result()
	if err == redis.Nil {
		cond.rp.RedisClient().Set(ctx, key, 1, 0)
	} else {
		count, _ := strconv.Atoi(val)
		count++
		cond.rp.RedisClient().Set(ctx, key, count, 0)

	}
	cond.lock.UnLock()
	cond.rp.Notify(key, "singal")
}

func (cond *Condition) Broadcast() {
	key := cond.conditionName + "-count"
	ctx := context.Background()
	cond.lock.Lock()
	cond.rp.RedisClient().Del(ctx, key)
	cond.lock.UnLock()
	cond.rp.Notify(key, "broadcast")
}

// 判断命名空间是否存在键值对
func NamespaceHasKeys(rdb *redis.Client, ctx context.Context, namespace string) bool {
	iter := rdb.Scan(ctx, 0, namespace+"*", 0).Iterator()
	for iter.Next(ctx) {
		return true
	}
	if err := iter.Err(); err != nil {
		return false
	}
	return false
}
