package span

import (
	"context"
	"errors"

	"github.com/opentracing/opentracing-go"
	"github.com/pojol/braid/lib/tracer"
	"github.com/uber/jaeger-client-go"
)

const (
	// EchoSpan echo span
	SystemCall = "braid.system.call.span"
)

// EchoTracer http request tracer
type SystemCallTracer struct {
	span    opentracing.Span
	tracing opentracing.Tracer
	parm    SpanParm

	starting bool
}

// CreateCallSpan
func CreateCallSpan(opts ...SpanOption) tracer.SpanFactory {
	return func(tracing interface{}) (tracer.ISpan, error) {

		t, ok := tracing.(opentracing.Tracer)
		if !ok {
			return nil, errors.New("")
		}

		parm := &SpanParm{nodeID: "unknown"}

		for _, v := range opts {
			v(parm)
		}

		et := &SystemCallTracer{
			tracing: t,
			parm:    *parm,
		}

		return et, nil
	}
}

// Begin starts the span and returns the updated context
func (t *SystemCallTracer) Begin(ctx interface{}) context.Context {
	mthonctx, ok := ctx.(context.Context)
	if !ok {
		// If the context is invalid, return a new background context
		return context.Background()
	}

	parentSpan := opentracing.SpanFromContext(mthonctx)
	if parentSpan != nil {
		t.span = t.tracing.StartSpan("System.call", opentracing.ChildOf(parentSpan.Context()))
	} else {
		t.span = t.tracing.StartSpan("System.call.root")
	}

	t.starting = true

	t.SetTag("node", t.parm.nodeID)

	// Create a new context with the span and return it
	return opentracing.ContextWithSpan(mthonctx, t.span)
}

func (t *SystemCallTracer) SetTag(key string, val interface{}) {
	if t.span != nil {
		t.span.SetTag(key, val)
	}
}

func (t *SystemCallTracer) GetID() string {
	if t.span != nil {
		if sc, ok := t.span.Context().(jaeger.SpanContext); ok {
			return sc.TraceID().String()
		}
	}

	return ""
}

// End finishes the span
func (t *SystemCallTracer) End(ctx interface{}) {

	if !t.starting {
		return
	}

	if t.span != nil {
		t.span.Finish()
	}
}
