package mgo

import (
	"braid/lib/tracer"
	"time"
)

type ConnInfo struct {
	Addr string
	Name string
}

type Parm struct {
	tracer tracer.ITracer

	connTimeout time.Duration
	poolSize    uint64

	conns []ConnInfo
}

// Option config wraps
type Option func(*Parm)

func WithTracer(trc tracer.ITracer) Option {
	return func(p *Parm) {
		p.tracer = trc
	}
}

func (p *Parm) _checkRepeat(addr string) bool {
	for _, c := range p.conns {
		if c.Addr == addr {
			return true
		}
	}

	return false
}

func WithConnTimeout(timeout time.Duration) Option {
	return func(p *Parm) {
		p.connTimeout = timeout
	}
}

func WithConnPoolSize(size uint64) Option {
	return func(p *Parm) {
		p.poolSize = size
	}
}

func AppendConn(info ConnInfo) Option {
	return func(p *Parm) {
		if !p._checkRepeat(info.Addr) {
			p.conns = append(p.conns, info)
		}
	}
}
