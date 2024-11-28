package actor

import (
	"sync/atomic"
	"time"
)

type TimerInfo struct {
	timer    *time.Timer
	dueTime  time.Duration
	interval time.Duration
	callback func(interface{}) error
	args     interface{}
	active   atomic.Bool  // 使用原子操作保证线程安全
	nextTick atomic.Value // 下次触发时间
}

func (t *TimerInfo) Stop() bool {
	if !t.active.Load() {
		return false
	}
	t.active.Store(false)
	return t.timer.Stop()
}

func (t *TimerInfo) Reset(interval time.Duration) bool {
	if interval > 0 {
		t.interval = interval
	}
	if t.interval <= 0 {
		return false
	}

	t.active.Store(true)
	t.nextTick.Store(time.Now().Add(t.interval))
	return t.timer.Reset(t.interval)
}

func (t *TimerInfo) IsActive() bool {
	return t.active.Load()
}

func (t *TimerInfo) Interval() time.Duration {
	return t.interval
}

func (t *TimerInfo) NextTrigger() time.Time {
	if v := t.nextTick.Load(); v != nil {
		return v.(time.Time)
	}
	return time.Time{}
}

func (t *TimerInfo) Execute() error {
	return t.callback(t.args)
}

func NewTimerInfo(dueTime, interval time.Duration, callback func(interface{}) error, args interface{}) *TimerInfo {
	t := &TimerInfo{
		dueTime:  dueTime,
		interval: interval,
		callback: callback,
		args:     args,
	}
	t.active.Store(true)

	if dueTime != 0 {
		t.nextTick.Store(time.Now().Add(dueTime))
	} else {
		t.nextTick.Store(time.Now().Add(interval))
	}

	t.timer = time.NewTimer(interval)
	return t
}
