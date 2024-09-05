package node

import "github.com/pojol/braid/core/workerthread"

type Parm struct {
	ID string // nod 的id全局唯一

	Ip   string // nod 的地址
	Port int    // nod 的端口号

	ClusterName string // 隶属于那个集群
	ServiceName string // 隶属于那个服务

	Sys workerthread.ISystem
}

type Option func(*Parm)

// tmp
func WithServiceInfo(ip string, port int) Option {
	return func(p *Parm) {
		p.Ip = ip
		p.Port = port
	}
}

func WithSystem(sys workerthread.ISystem) Option {
	return func(p *Parm) {
		p.Sys = sys
	}
}
