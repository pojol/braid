package node

import (
	"github.com/pojol/braid/lib/tracer"
)

type SystemParm struct {
	NodeID string
	Ip     string
	Port   int
	Tracer tracer.ITracer
}

type SystemOption func(*SystemParm)

func SystemWithAcceptor(port int) SystemOption {
	return func(sp *SystemParm) {
		sp.Port = port
	}
}

func SystemWithTracer(t tracer.ITracer) SystemOption {
	return func(sp *SystemParm) {
		sp.Tracer = t
	}
}
