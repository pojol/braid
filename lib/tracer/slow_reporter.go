package tracer

import (
	"sync"
	"sync/atomic"
	"time"

	"github.com/uber/jaeger-client-go"
)

type reporterQueueItemType int

const (
	defaultQueueSize           = 100
	defaultBufferFlushInterval = 1 * time.Second

	reporterQueueItemSpan reporterQueueItemType = iota
	reporterQueueItemClose
)

// reporterStats implements reporterstats.ReporterStats.
type reporterStats struct {
	droppedCount int64 // provided to Transports to report data loss to the backend
}

// SpansDroppedFromQueue implements reporterstats.ReporterStats.
func (r *reporterStats) SpansDroppedFromQueue() int64 {
	return atomic.LoadInt64(&r.droppedCount)
}

func (r *reporterStats) incDroppedCount() {
	atomic.AddInt64(&r.droppedCount, 1)
}

type reporterQueueItem struct {
	itemType reporterQueueItemType
	span     *jaeger.Span
	close    *sync.WaitGroup
}

type slowReporter struct {
	// These fields must be first in the struct because `sync/atomic` expects 64-bit alignment.
	// Cf. https://github.com/uber/jaeger-client-go/issues/155, https://goo.gl/zW7dgq
	queueLength int64 // used to update metrics.Gauge
	closed      int64 // 0 - not closed, 1 - closed

	metrics             *jaeger.Metrics
	queueSize           int
	bufferFlushInterval time.Duration

	sampler *jaeger.ProbabilisticSampler

	sender        jaeger.Transport
	queue         chan reporterQueueItem
	reporterStats *reporterStats
}

// NewSlowReporter creates a new reporter that sends spans out of process by means of Sender.
func NewSlowReporter(sender jaeger.Transport, metrics *jaeger.Metrics, probabilistic float64) jaeger.Reporter {

	if metrics == nil {
		metrics = jaeger.NewNullMetrics()
	}

	sampler, _ := jaeger.NewProbabilisticSampler(probabilistic)

	reporter := &slowReporter{
		sender:              sender,
		queueSize:           defaultQueueSize,
		bufferFlushInterval: defaultBufferFlushInterval,
		metrics:             metrics,
		sampler:             sampler,
		queue:               make(chan reporterQueueItem, defaultQueueSize),
		reporterStats:       new(reporterStats),
	}
	/*
		if receiver, ok := sender.(reporterstats.Receiver); ok {
			receiver.SetReporterStats(reporter.reporterStats)
		}
	*/
	go reporter.processQueue()
	return reporter
}

// Report implements Report() method of Reporter.
func (r *slowReporter) Report(span *jaeger.Span) {

	select {
	// Need to retain the span otherwise it will be released
	case r.queue <- reporterQueueItem{itemType: reporterQueueItemSpan, span: span.Retain()}:
		atomic.AddInt64(&r.queueLength, 1)
	default:
		r.metrics.ReporterDropped.Inc(1)
		r.reporterStats.incDroppedCount()
	}
}

// Close implements Close() method of Reporter by waiting for the queue to be drained.
func (r *slowReporter) Close() {
	if swapped := atomic.CompareAndSwapInt64(&r.closed, 0, 1); !swapped {
		return
	}
	r.sendCloseEvent()
	_ = r.sender.Close()
}

func (r *slowReporter) sendCloseEvent() {
	wg := &sync.WaitGroup{}
	wg.Add(1)
	item := reporterQueueItem{itemType: reporterQueueItemClose, close: wg}

	r.queue <- item // if the queue is full we will block until there is space
	atomic.AddInt64(&r.queueLength, 1)
	wg.Wait()
}

// processQueue reads spans from the queue, converts them to Thrift, and stores them in an internal buffer.
// When the buffer length reaches batchSize, it is flushed by submitting the accumulated spans to Jaeger.
// Buffer also gets flushed automatically every batchFlushInterval seconds, just in case the tracer stopped
// reporting new spans.
func (r *slowReporter) processQueue() {
	// flush causes the Sender to flush its accumulated spans and clear the buffer
	flush := func() {
		if flushed, err := r.sender.Flush(); err != nil {
			r.metrics.ReporterFailure.Inc(int64(flushed))
		} else if flushed > 0 {
			r.metrics.ReporterSuccess.Inc(int64(flushed))
		}
	}

	timer := time.NewTicker(r.bufferFlushInterval)
	for {
		select {
		case <-timer.C:
			flush()
		case item := <-r.queue:
			atomic.AddInt64(&r.queueLength, -1)

			switch item.itemType {
			case reporterQueueItemSpan:
				span := item.span
				var sample bool

				//if span.Duration() > (time.Millisecond*tracer.cfg.SlowSpan) && span.OperationName() != optionHTTPRequest {
				//log.SysSlow(span.OperationName(), span.SpanContext().TraceID().String(), int(span.Duration().Milliseconds()), "slow span")
				//	sample = true
				//} else {
				sample, _ = r.sampler.IsSampled(span.SpanContext().TraceID(), span.OperationName())
				//}

				if sample {
					if flushed, err := r.sender.Append(span); err != nil {
						r.metrics.ReporterFailure.Inc(int64(flushed))
					} else if flushed > 0 {
						r.metrics.ReporterSuccess.Inc(int64(flushed))
						// to reduce the number of gauge stats, we only emit queue length on flush
						r.metrics.ReporterQueueLength.Update(atomic.LoadInt64(&r.queueLength))
					}
				}

				span.Release()
			case reporterQueueItemClose:
				timer.Stop()
				flush()
				item.close.Done()
				return
			}
		}
	}
}
