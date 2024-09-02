package tracer

import (
	"errors"
	"fmt"
	"io"
	"time"

	opentracing "github.com/opentracing/opentracing-go"
	"github.com/uber/jaeger-client-go"
	jaegerCfg "github.com/uber/jaeger-client-go/config"
	"github.com/uber/jaeger-client-go/transport"
	"github.com/uber/jaeger-lib/metrics"
)

const (
	// Name module name
	Name = "JaegerTracing"
)

var (
	// ErrFactoryNotExist factory not exist
	ErrFactoryNotExist = errors.New("factory not exist")
)

func newTransport(rc *jaegerCfg.ReporterConfig) (jaeger.Transport, error) {
	switch {
	case rc.CollectorEndpoint != "":
		httpOptions := []transport.HTTPOption{transport.HTTPBatchSize(1), transport.HTTPHeaders(rc.HTTPHeaders)}
		if rc.User != "" && rc.Password != "" {
			httpOptions = append(httpOptions, transport.HTTPBasicAuth(rc.User, rc.Password))
		}
		return transport.NewHTTPTransport(rc.CollectorEndpoint, httpOptions...), nil
	default:
		return jaeger.NewUDPTransport(rc.LocalAgentHostPort, 0)
	}
}

func BuildWithOption(opts ...Option) ITracer {

	p := Parm{
		Probabilistic: 1,
		SlowRequest:   time.Millisecond * 200,
		SlowSpan:      time.Millisecond * 50,
		Name:          "braid",
		SamplerType:   jaeger.SamplerTypeConst,
	}

	for _, opt := range opts {
		opt(&p)
	}

	jcfg := jaegerCfg.Configuration{
		Sampler: &jaegerCfg.SamplerConfig{
			Type:  p.SamplerType,
			Param: p.Probabilistic,
		},
		Reporter: &jaegerCfg.ReporterConfig{
			LogSpans:           true,
			CollectorEndpoint:  p.CollectorEndpoint, //with http
			LocalAgentHostPort: p.LocalAgentHostPort,
		},
		ServiceName: p.Name,
	}

	jt := &jaegerTracing{
		parm:        p,
		serviceName: p.Name,
		jcfg:        jcfg,
		factory:     make(map[string]SpanFactory),
	}

	for _, v := range p.ImportFactory {
		if _, ok := jt.factory[v.Name]; !ok {
			jt.factory[v.Name] = v.Factory
		}
	}

	sender, err := newTransport(jt.jcfg.Reporter)
	if err != nil {
		panic(fmt.Errorf("%v Dependency check error %v [%v]", jt.serviceName, "jaegertracing", err.Error()))
	}

	r := jaegerCfg.Reporter(NewSlowReporter(sender, nil, jt.parm.Probabilistic))
	m := jaegerCfg.Metrics(metrics.NullFactory)

	jtracing, closer, err := jt.jcfg.NewTracer(r, m)
	if err != nil {
		panic(fmt.Errorf("%v Dependency check error %v [%v]", jt.serviceName, "jaegertracing", err.Error()))
	}

	jt.tracing = jtracing
	jt.closer = closer

	return jt
}

func (jt *jaegerTracing) Init() error {

	return nil
}

type jaegerTracing struct {
	parm        Parm
	serviceName string
	jcfg        jaegerCfg.Configuration

	closer  io.Closer
	tracing opentracing.Tracer

	factory map[string]SpanFactory
}

func (jt *jaegerTracing) Run() {

}

func (jt *jaegerTracing) GetSpan(strategy string) (ISpan, error) {

	spanfactory, ok := jt.factory[strategy]
	if !ok {
		return nil, ErrFactoryNotExist
	}

	span, err := spanfactory(jt.tracing)
	if err != nil {
		return nil, err
	}

	return span, nil
}

func (jt *jaegerTracing) GetTracing() interface{} {
	return jt.tracing
}

func (jt *jaegerTracing) Close() {
	jt.closer.Close()
}
