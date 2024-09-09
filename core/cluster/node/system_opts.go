package node

import "github.com/pojol/braid/core"

type SystemParm struct {
	ServiceName string
	NodeID      string
	Ip          string
	Port        int

	Constructors []ActorConstructor
}

type ActorConstructor struct {
	Type        string
	Constructor core.CreateFunc
}

type SystemOption func(*SystemParm)

func SystemService(serviceName, nodeID string) SystemOption {
	return func(sp *SystemParm) {
		sp.NodeID = nodeID
		sp.ServiceName = serviceName
	}
}

func SystemActorConstructor(lst []ActorConstructor) SystemOption {
	return func(sp *SystemParm) {
		sp.Constructors = append(sp.Constructors, lst...)
	}
}

func SystemWithAcceptor(port int) SystemOption {
	return func(sp *SystemParm) {
		sp.Port = port
	}
}
