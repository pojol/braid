package benchmarktimer

import (
	"fmt"
	"sync/atomic"
	"testing"
	"time"

	"github.com/pojol/braid/lib/timewheel"
	"golang.org/x/exp/rand"
)

type timerActor struct {
	timerCount   int
	sum          int64
	triggerCount int64
	updateCount  int64

	tw       *timewheel.TimeWheel
	lastTick time.Time
}

func (a *timerActor) RegisterTimer(dueTime int64, interval int64, f func() error, args interface{}) *timewheel.Timer {
	return a.tw.AddTimer(
		time.Duration(dueTime)*time.Millisecond,
		time.Duration(interval)*time.Millisecond,
		f,
		args,
	)
}

func (a *timerActor) RemoveTimer(t *timewheel.Timer) {
	a.tw.RemoveTimer(t)
}

func (a *timerActor) Init() {

	for i := 0; i < a.timerCount; i++ {
		a.RegisterTimer(0, int64(rand.Intn(100)+1), func() error {
			// 模拟固定计算量
			for j := 0; j < 1000; j++ {
				atomic.AddInt64(&a.sum, int64(j))
			}
			atomic.AddInt64(&a.triggerCount, 1)
			return nil
		}, nil)
	}
}

func (a *timerActor) Update() {
	now := time.Now()
	if now.Sub(a.lastTick) >= a.tw.Interval() {
		a.tw.Tick()
		a.lastTick = now
	}
	atomic.AddInt64(&a.updateCount, 1)
}

func BenchmarkActorTimer(b *testing.B) {
	timerCounts := []int{1000, 10000, 100000}
	testDuration := 10 * time.Second

	for _, count := range timerCounts {
		b.Run(fmt.Sprintf("Timers_%d", count), func(b *testing.B) {

			actor := &timerActor{
				tw:         timewheel.New(10*time.Millisecond, 100), // 100个槽位，每个槽位10ms
				timerCount: count,
				lastTick:   time.Now(),
			}
			actor.Init()

			b.ResetTimer()
			start := time.Now()
			for time.Since(start) < testDuration {
				actor.Update()
			}
			duration := time.Since(start)

			b.StopTimer()

			updatesPerSecond := float64(actor.updateCount) / duration.Seconds()
			triggersPerSecond := float64(actor.triggerCount) / duration.Seconds()
			averageUpdateDuration := duration.Seconds() / float64(actor.updateCount) * 1e6 // in microseconds

			b.ReportMetric(updatesPerSecond, "updates/sec")
			b.ReportMetric(triggersPerSecond, "triggers/sec")
			b.ReportMetric(averageUpdateDuration, "µs/update")

			fmt.Printf("Timers: %d, Updates: %d, Triggers: %d, Avg Update Duration: %.3f µs\n",
				count, actor.updateCount, actor.triggerCount, averageUpdateDuration)
		})
	}
}
