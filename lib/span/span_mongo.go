package span

import (
	"context"
	"errors"

	"github.com/opentracing/opentracing-go"
	"github.com/pojol/braid/lib/tracer"
	"github.com/uber/jaeger-client-go"
)

// MethonTracer methon tracer
type MongoTracer struct {
	span    opentracing.Span
	tracing opentracing.Tracer

	starting bool
}

const (
	MongoSpan = "tracer_span_mongo"
)

func CreateMongoSpanFactory() tracer.SpanFactory {
	return func(tracing interface{}) (tracer.ISpan, error) {

		t, ok := tracing.(opentracing.Tracer)
		if !ok {
			return nil, errors.New("")
		}

		rt := &MongoTracer{
			tracing: t,
		}

		return rt, nil
	}
}

// Begin 开始监听
func (r *MongoTracer) Begin(ctx interface{}) {

	mthonctx, ok := ctx.(context.Context)
	if !ok {
		return
	}

	parentSpan := opentracing.SpanFromContext(mthonctx)
	if parentSpan != nil {
		r.span = r.tracing.StartSpan("MongoSpan", opentracing.ChildOf(parentSpan.Context()))
	}

	r.starting = true
}

func (r *MongoTracer) SetTag(key string, val interface{}) {
	if r.span != nil {
		r.span.SetTag(key, val)
	}
}

func (r *MongoTracer) GetID() string {
	if r.span != nil {
		if sc, ok := r.span.Context().(jaeger.SpanContext); ok {
			return sc.TraceID().String()
		}
	}

	return ""
}

// End 结束监听
func (r *MongoTracer) End(ctx interface{}) {

	if !r.starting {
		return
	}

	if r.span != nil {
		r.span.Finish()
	}

}
