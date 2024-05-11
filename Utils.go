package stormi

import (
	"time"
)

type timer struct {
	time time.Time
}

type utils struct{}

var Utils utils

func (utils) NewTimer() timer {
	t := timer{}
	t.time = time.Now()
	return t
}

func (tt timer) Stamp() time.Duration {
	return time.Since(tt.time)
}

func (tt *timer) StampAndReset() time.Duration {
	t := time.Since(tt.time)
	tt.time = time.Now()
	return t
}
