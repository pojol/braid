package timewheel

import (
	"container/list"
	"sync"
	"time"
)

type Timer struct {
	delay    time.Duration
	nextTick time.Time
	interval time.Duration
	f        func(interface{}) error
	args     interface{}
}

type TimeWheel struct {
	interval   time.Duration
	slots      []*list.List
	currentPos int
	slotNum    int

	mutex    sync.Mutex
	shutdown bool
}

func New(interval time.Duration, slotNum int) *TimeWheel {
	tw := &TimeWheel{
		interval:   interval,
		slots:      make([]*list.List, slotNum),
		currentPos: 0,
		slotNum:    slotNum,
		shutdown:   false,
	}

	for i := 0; i < slotNum; i++ {
		tw.slots[i] = list.New()
	}

	return tw
}

func (tw *TimeWheel) AddTimer(delay time.Duration, interval time.Duration, f func(interface{}) error, args interface{}) *Timer {
	tw.mutex.Lock()
	defer tw.mutex.Unlock()

	if tw.shutdown {
		return nil
	}

	if interval == 0 {
		// 处理间隔为0的错误情况
		panic("TimeWheel interval cannot be zero")
	}

	timer := &Timer{
		delay:    delay,
		nextTick: time.Now().Add(delay),
		interval: interval,
		f:        f,
		args:     args,
	}

	tw.addTimer(timer)
	return timer
}

func (tw *TimeWheel) addTimer(t *Timer) {
	pos, _ := tw.getPositionAndCircle(t.delay)
	tw.slots[pos].PushBack(t)
}

func (tw *TimeWheel) RemoveTimer(t *Timer) {
	tw.mutex.Lock()
	defer tw.mutex.Unlock()

	if tw.shutdown {
		return
	}

	for _, slot := range tw.slots {
		for e := slot.Front(); e != nil; e = e.Next() {
			if e.Value == t {
				slot.Remove(e)
				return
			}
		}
	}
}

func (tw *TimeWheel) getPositionAndCircle(d time.Duration) (pos int, circle int) {
	if tw.interval < time.Microsecond {
		tw.interval = time.Microsecond // 设置一个最小间隔
	}

	totalSlots := int(d / tw.interval)
	if totalSlots == 0 {
		totalSlots = 1
	}

	circle = totalSlots / tw.slotNum
	pos = (tw.currentPos + totalSlots) % tw.slotNum

	return
}

func (tw *TimeWheel) Interval() time.Duration {
	return tw.interval
}

func (tw *TimeWheel) Tick() {
	tw.mutex.Lock()
	defer tw.mutex.Unlock()

	if tw.shutdown {
		return
	}

	currentSlot := tw.slots[tw.currentPos]
	now := time.Now()

	for e := currentSlot.Front(); e != nil; {
		timer := e.Value.(*Timer)
		if now.After(timer.nextTick) || now.Equal(timer.nextTick) {
			timer.f(timer.args)
			next := e.Next()
			currentSlot.Remove(e)
			e = next

			if timer.interval > 0 {
				for timer.nextTick.Before(now) || timer.nextTick.Equal(now) {
					timer.nextTick = timer.nextTick.Add(timer.interval)
				}
				tw.addTimer(timer)
			}
		} else {
			e = e.Next()
		}
	}

	tw.currentPos = (tw.currentPos + 1) % tw.slotNum
}

func (tw *TimeWheel) Shutdown() {
	tw.mutex.Lock()
	defer tw.mutex.Unlock()

	if tw.shutdown {
		return
	}

	tw.shutdown = true

	// 清空所有槽位
	for i := 0; i < tw.slotNum; i++ {
		tw.slots[i].Init()
	}

	// 重置当前位置
	tw.currentPos = 0
}
