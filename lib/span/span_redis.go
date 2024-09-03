package span

import (
	"context"
	"errors"

	"github.com/opentracing/opentracing-go"
	"github.com/opentracing/opentracing-go/ext"
	"github.com/pojol/braid/lib/tracer"
	"github.com/uber/jaeger-client-go"
)

const (
	// RedisSpan redis span
	RedisSpan = "tracer_span_redis"
)

// RedisTracer redis tracer
type RedisTracer struct {
	span    opentracing.Span
	tracing opentracing.Tracer

	starting bool

	Cmd string
}

func CreateRedisSpanFactory() tracer.SpanFactory {
	return func(tracing interface{}) (tracer.ISpan, error) {

		t, ok := tracing.(opentracing.Tracer)
		if !ok {
			return nil, errors.New("")
		}

		rt := &RedisTracer{
			tracing: t,
		}

		return rt, nil
	}
}

// Begin 开始监听
func (r *RedisTracer) Begin(ctx interface{}) {

	redisctx, ok := ctx.(context.Context)
	if !ok {
		return
	}

	parentSpan := opentracing.SpanFromContext(redisctx)
	if parentSpan != nil {
		r.span = r.tracing.StartSpan(r.Cmd, opentracing.ChildOf(parentSpan.Context()))

		ext.DBType.Set(r.span, "Redis")
	}

	r.starting = true
}

func (r *RedisTracer) SetTag(key string, val interface{}) {
	if r.span != nil {
		r.span.SetTag(key, val)
	}
}

func (r *RedisTracer) GetID() string {
	if r.span != nil {
		if sc, ok := r.span.Context().(jaeger.SpanContext); ok {
			return sc.TraceID().String()
		}
	}

	return ""
}

// End 结束监听
func (r *RedisTracer) End(ctx interface{}) {

	if !r.starting {
		return
	}

	if r.span != nil {
		r.span.Finish()
	}

}
