package span

import (
	"context"
	"errors"
	"net/http"
	"time"

	"github.com/labstack/echo/v4"
	"github.com/opentracing/opentracing-go"
	"github.com/opentracing/opentracing-go/ext"
	"github.com/pojol/braid/lib/tracer"
	"github.com/uber/jaeger-client-go"
)

// TraceKey 主键类型
type TraceKey int

const (
	// RequestKey 请求的键值
	RequestKey TraceKey = 1000 + iota
)

const (
	// EchoSpan echo span
	EchoSpan = "tracer_span_echo"
)

// EchoTracer http request tracer
type EchoTracer struct {
	span    opentracing.Span
	tracing opentracing.Tracer

	starting bool

	requestID string
	beginTime time.Time
}

// createEchoTraceSpan 构建 echo tracer
func CreateEchoTraceSpan() tracer.SpanFactory {
	return func(tracing interface{}) (tracer.ISpan, error) {

		t, ok := tracing.(opentracing.Tracer)
		if !ok {
			return nil, errors.New("")
		}

		et := &EchoTracer{
			tracing: t,
		}

		return et, nil
	}
}

// Begin 开始捕捉
func (t *EchoTracer) Begin(ctx interface{}) {

	echoContext, ok := ctx.(echo.Context)
	if !ok {
		return
	}

	req := echoContext.Request()
	t.beginTime = time.Now()

	if ctx, err := t.tracing.Extract(opentracing.HTTPHeaders,
		opentracing.HTTPHeadersCarrier(req.Header)); err != nil {
		t.span = t.tracing.StartSpan("HttpRequest")
	} else {
		t.span = t.tracing.StartSpan("HttpRequest", ext.RPCServerOption(ctx))
	}

	ext.HTTPMethod.Set(t.span, req.Method)
	ext.HTTPUrl.Set(t.span, req.URL.String())
	rc := opentracing.ContextWithSpan(req.Context(), t.span)
	// Set request ID for context.
	if sc, ok := t.span.Context().(jaeger.SpanContext); ok {
		t.requestID = sc.TraceID().String()
		rc = context.WithValue(rc, RequestKey, t.requestID)
	}

	req = req.WithContext(opentracing.ContextWithSpan(rc, t.span))
	echoContext.SetRequest(req)

	t.starting = true
}

func (t *EchoTracer) SetTag(key string, val interface{}) {
	if t.span != nil {
		t.span.SetTag(key, val)
	}
}

func (t *EchoTracer) GetID() string {
	if t.span != nil {
		if sc, ok := t.span.Context().(jaeger.SpanContext); ok {
			return sc.TraceID().String()
		}
	}

	return ""
}

// End 结束捕捉
func (t *EchoTracer) End(ctx interface{}) {

	if !t.starting {
		return
	}

	ectx, ok := ctx.(echo.Context)
	if !ok {
		return
	}

	status := ectx.Response().Status
	committed := ectx.Response().Committed

	ext.HTTPStatusCode.Set(t.span, uint16(status))
	if status >= http.StatusInternalServerError || !committed {
		ext.Error.Set(t.span, true)
	}

	//executionTime := time.Now().Sub(t.beginTime)
	//if executionTime > tracer.cfg.SlowRequest {
	//log.SysSlow(ectx.Path(), t.requestID, int(executionTime), "slow request")
	//}

	t.span.Finish()
}
