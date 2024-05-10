package stormi

import (
	"time"
)

type ticker struct {
	time time.Time
}

type utils struct{}

var Utils utils

func (utils) NewTimer() ticker {
	t := ticker{
		time: time.Now(),
	}
	return t
}

func (tt ticker) Stamp() time.Duration {
	return time.Since(tt.time)
}

func (tt *ticker) StampAndReset() time.Duration {
	t := time.Since(tt.time)
	tt.time = time.Now()
	return t
}
