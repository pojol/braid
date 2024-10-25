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
	ID     string // node's globally unique ID
	Weight int

	Ip   string
	Port int

	Tracer tracer.ITracer

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

func NodeWithIP(ip string) NodeOption {
	return func(np *NodeParm) {
		np.Ip = ip
	}
}

func NodeWithPort(port int) NodeOption {
	return func(np *NodeParm) {
		np.Port = port
	}
}

func NodeWithTracer(t tracer.ITracer) NodeOption {
	return func(np *NodeParm) {
		np.Tracer = t
	}
}
