package core

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
	ID string // nod 的id全局唯一

	Ip   string // nod 的地址
	Port int    // nod 的端口号

	ClusterName string // 隶属于那个集群
	ServiceName string // 隶属于那个服务

	Sys ISystem
}

type NodeOption func(*NodeParm)

// tmp
func WithServiceInfo(ip string, port int) NodeOption {
	return func(p *NodeParm) {
		p.Ip = ip
		p.Port = port
	}
}

func WithNodeID(id string) NodeOption {
	return func(np *NodeParm) {
		np.ID = id
	}
}

func WithSystem(sys ISystem) NodeOption {
	return func(p *NodeParm) {
		p.Sys = sys
	}
}
