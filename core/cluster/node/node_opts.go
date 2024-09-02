package node

type Parm struct {
	ID   string // nod 的id全局唯一
	Name string // nod 的名称

	Addr string // nod 的地址
	Port int    // nod 的端口号

	ClusterName string // 隶属于那个集群
	ServiceName string // 隶属于那个服务
}

type Option func(*Parm)
