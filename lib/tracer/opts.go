package tracer

import "time"

type TracerFactory struct {
	Name    string
	Factory SpanFactory
}

// Parm https://github.com/jaegertracing/jaeger-client-go/blob/master/config/config.go
type Parm struct {
	Name               string
	CollectorEndpoint  string  // jaeger http地址
	LocalAgentHostPort string  // jaeger udp地址
	Probabilistic      float64 // 采样率
	SamplerType        string
	SlowRequest        time.Duration // 一旦request超出设置的SlowRequest（ms）时间，则一定会有一条slow日志
	SlowSpan           time.Duration // 一旦span超出设置的SlowSpan（ms）时间，则一定会有一条slow日志
	ImportFactory      []TracerFactory
}

// Option config wraps
type Option func(*Parm)

func WithServiceName(name string) Option {
	return func(c *Parm) {
		c.Name = name
	}
}

// WithProbabilistic 采样率 0 ~ 1
func WithProbabilistic(probabilistic float64) Option {
	return func(c *Parm) {
		c.Probabilistic = probabilistic
	}
}

// WithSamplerType 采样类型
func WithSamplerType(samplertype string) Option {
	return func(c *Parm) {
		c.SamplerType = samplertype
	}
}

// WithSlowRequest 单次请求阀值（超过会有慢日志
func WithSlowRequest(ms int) Option {
	return func(c *Parm) {
		c.SlowRequest = time.Duration(ms) * time.Millisecond
	}
}

// WithSlowSpan 单次调用阀值（超过会有慢日志
func WithSlowSpan(ms int) Option {
	return func(c *Parm) {
		c.SlowSpan = time.Duration(ms) * time.Millisecond
	}
}

// WithHTTP http://172.17.0.1:14268/api/traces
func WithHTTP(CollectorEndpoint string) Option {
	return func(c *Parm) {
		c.CollectorEndpoint = CollectorEndpoint
	}
}

// WithUDP 172.17.0.1:6831
func WithUDP(LocalAgentHostPort string) Option {
	return func(c *Parm) {
		c.LocalAgentHostPort = LocalAgentHostPort
	}
}

func WithSpanFactory(factory ...TracerFactory) Option {
	return func(c *Parm) {
		for _, v := range factory {
			c.ImportFactory = append(c.ImportFactory, v)
		}
	}
}
