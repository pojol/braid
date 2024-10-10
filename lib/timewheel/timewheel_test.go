package timewheel

import (
	"context"
	"sync"
	"testing"
	"time"
)

func TestNew(t *testing.T) {
	interval := time.Millisecond * 100
	slotNum := 10
	tw := New(interval, slotNum)

	if tw.interval != interval {
		t.Errorf("Expected interval %v, got %v", interval, tw.interval)
	}

	if tw.slotNum != slotNum {
		t.Errorf("Expected slotNum %d, got %d", slotNum, tw.slotNum)
	}

	if len(tw.slots) != slotNum {
		t.Errorf("Expected %d slots, got %d", slotNum, len(tw.slots))
	}
}

func TestAddTimer(t *testing.T) {
	tw := New(time.Millisecond*100, 10)

	delay := time.Millisecond * 200
	interval := time.Millisecond * 100
	timer := tw.AddTimer(delay, interval, func(interface{}) error {
		return nil
	}, nil)

	if timer == nil {
		t.Error("Expected timer to be created, got nil")
	}

	if timer.delay != delay {
		t.Errorf("Expected delay %v, got %v", delay, timer.delay)
	}

	if timer.interval != interval {
		t.Errorf("Expected interval %v, got %v", interval, timer.interval)
	}
}

func TestRemoveTimer(t *testing.T) {
	tw := New(time.Millisecond*100, 10)

	timer := tw.AddTimer(time.Second, time.Second, func(interface{}) error { return nil }, nil)
	tw.RemoveTimer(timer)

	// Check if the timer was removed from all slots
	for _, slot := range tw.slots {
		for e := slot.Front(); e != nil; e = e.Next() {
			if e.Value == timer {
				t.Error("Timer was not removed from the slot")
			}
		}
	}
}

func TestTick(t *testing.T) {
	tw := New(time.Millisecond*100, 10)

	var wg sync.WaitGroup
	wg.Add(1)

	executed := false
	tw.AddTimer(time.Millisecond*50, 100, func(interface{}) error {
		executed = true
		wg.Done()
		return nil
	}, nil)

	// Simulate ticks
	for i := 0; i < 2; i++ {
		tw.Tick()
		time.Sleep(time.Millisecond * 60)
	}

	// Add timeout to prevent test from hanging
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()

	done := make(chan struct{})
	go func() {
		wg.Wait()
		close(done)
	}()

	select {
	case <-done:
		if !executed {
			t.Error("Timer function was not executed")
		}
	case <-ctx.Done():
		t.Error("Test timed out")
	}
}

func TestRecurringTimer(t *testing.T) {
	tw := New(time.Millisecond*100, 10)

	count := 0
	var mu sync.Mutex
	expectedExecutions := 2

	timer := tw.AddTimer(time.Millisecond*50, time.Millisecond*200, func(interface{}) error {
		mu.Lock()
		defer mu.Unlock()
		count++
		t.Logf("Timer executed at %v, count: %d", time.Now(), count)
		return nil
	}, nil)

	// Run the TimeWheel for a fixed duration
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()

	go func() {
		ticker := time.NewTicker(tw.interval)
		defer ticker.Stop()
		for {
			select {
			case <-ctx.Done():
				return
			case <-ticker.C:
				tw.Tick()
			}
		}
	}()

	// Wait for the test duration
	<-ctx.Done()

	// Stop the timer to prevent further executions
	tw.RemoveTimer(timer)

	mu.Lock()
	finalCount := count
	mu.Unlock()

	if finalCount != expectedExecutions {
		t.Errorf("Expected recurring timer to execute %d times, but it executed %d times", expectedExecutions, finalCount)
	}
}

func TestShutdown(t *testing.T) {
	tw := New(time.Millisecond*100, 10)

	tw.AddTimer(time.Second, time.Second, func(interface{}) error { return nil }, nil)

	tw.Shutdown()

	if !tw.shutdown {
		t.Error("TimeWheel should be in shutdown state")
	}

	// Check if all slots are empty
	for _, slot := range tw.slots {
		if slot.Len() != 0 {
			t.Error("All slots should be empty after shutdown")
		}
	}

	// Try to add a new timer after shutdown
	timer := tw.AddTimer(time.Second, time.Second, func(interface{}) error { return nil }, nil)
	if timer != nil {
		t.Error("Should not be able to add timer after shutdown")
	}
}

func TestConcurrency(t *testing.T) {
	tw := New(time.Millisecond*10, 10)

	var wg sync.WaitGroup
	for i := 0; i < 100; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			tw.AddTimer(time.Millisecond*50, 500, func(interface{}) error { return nil }, nil)
		}()
	}

	// Concurrent ticks
	for i := 0; i < 10; i++ {
		go tw.Tick()
	}

	// Add timeout to prevent test from hanging
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	done := make(chan struct{})
	go func() {
		wg.Wait()
		close(done)
	}()

	select {
	case <-done:
		// Test completed successfully
	case <-ctx.Done():
		t.Error("Test timed out")
	}
}

func TestZeroIntervalPanic(t *testing.T) {
	tw := New(time.Millisecond*100, 10)

	defer func() {
		if r := recover(); r == nil {
			t.Error("The code did not panic on zero interval")
		}
	}()

	tw.AddTimer(time.Second, 0, func(interface{}) error { return nil }, nil)
}
