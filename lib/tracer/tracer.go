package tracer

import "context"

// SpanFactory span 工厂
type SpanFactory func(interface{}) (ISpan, error)

// ISpan span interface
type ISpan interface {
	Begin(ctx interface{}) context.Context
	SetTag(key string, val interface{})
	GetID() string
	End(ctx interface{})
}

// ITracer tracer interface
type ITracer interface {
	GetSpan(strategy string) (ISpan, error)

	GetTracing() interface{}
}
