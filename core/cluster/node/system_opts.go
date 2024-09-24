package node

type SystemParm struct {
	ServiceName string
	NodeID      string
	Ip          string
	Port        int
}

type SystemOption func(*SystemParm)

func SystemService(serviceName, nodeID string) SystemOption {
	return func(sp *SystemParm) {
		sp.NodeID = nodeID
		sp.ServiceName = serviceName
	}
}

func SystemWithAcceptor(port int) SystemOption {
	return func(sp *SystemParm) {
		sp.Port = port
	}
}
