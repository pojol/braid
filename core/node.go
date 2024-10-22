package core

import "github.com/pojol/braid/lib/tracer"

/*
	init - 初始化进程
	update - 将一堆执行线程丢到node的运行时驱动
	close - 监听退出信号，通知到各执行线程停止接受新处理，等待当前处理结束退出
*/

type INode interface {
	Init(...NodeOption) error
	Update()
	WaitClose()

	ID() string
	System() ISystem
}

type NodeParm struct {
	ID     string // nod 的id全局唯一
	Weight int

	Ip   string // nod 的地址
	Port int    // nod 的端口号

	SystemOpts []SystemOption

	Loader  IActorLoader
	Factory IActorFactory
}

type NodeOption func(*NodeParm)

// tmp
func NodeWithServiceInfo(ip string, port int) NodeOption {
	return func(p *NodeParm) {
		p.Ip = ip
		p.Port = port
	}
}

func NodeWithID(id string) NodeOption {
	return func(np *NodeParm) {
		np.ID = id
	}
}

func NodeWithWeight(weight int) NodeOption {
	return func(np *NodeParm) {
		np.Weight = weight
	}
}

func NodeWithLoader(load IActorLoader) NodeOption {
	return func(p *NodeParm) {
		p.Loader = load
	}
}

func NodeWithFactory(factory IActorFactory) NodeOption {
	return func(np *NodeParm) {
		np.Factory = factory
	}
}

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

func NodeWithSystemOpts(opts ...SystemOption) NodeOption {
	return func(np *NodeParm) {
		np.SystemOpts = append(np.SystemOpts, opts...)
	}
}
