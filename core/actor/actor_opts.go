package actor

type SystemParm struct {
	ServiceName string
	NodeID      string

	Constructors []ActorConstructor
}

type ActorConstructor struct {
	Type        string
	Constructor CreateFunc
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

type CreateActorParm struct {
	ID     string
	InsPtr interface{}
}

type CreateActorOption func(*CreateActorParm)

func CreateActorWithID(id string) CreateActorOption {
	return func(p *CreateActorParm) {
		p.ID = id
	}
}

func CreateActorWithIns(ins interface{}) CreateActorOption {
	return func(p *CreateActorParm) {
		p.InsPtr = ins
	}
}
