package node

type SystemParm struct {
	NodeID string
	Ip     string
	Port   int
}

type SystemOption func(*SystemParm)

func SystemWithAcceptor(port int) SystemOption {
	return func(sp *SystemParm) {
		sp.Port = port
	}
}
