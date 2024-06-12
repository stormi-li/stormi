package stormi

//实现分布式WaitGroup，实现分布式NewCond，实现分布式读写锁

type SyncProxy struct {
	rp *RedisProxy
}

func NewSyncProxy(rp *RedisProxy) *SyncProxy {
	sp := SyncProxy{}
	sp.rp = rp
	return &sp
}

