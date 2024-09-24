package test

import (
	"github.com/pojol/braid/core"
)

// MockActorFactory is a factory for creating actors
type MockActorFactory struct {
	constructors map[string]*core.ActorConstructor
}

// NewActorFactory create new actor factory
func BuildActorFactory() *MockActorFactory {
	factory := &MockActorFactory{
		constructors: make(map[string]*core.ActorConstructor),
	}

	factory.bind("MockUserActor", core.ActorRegisteraionType_Dynamic, 80, NewUserActor)
	factory.bind("MockClacActor", core.ActorRegisteraionType_Dynamic, 20, NewClacActor)

	return factory
}

// Bind associates an actor type with its constructor function
func (factory *MockActorFactory) bind(actorType string, regType string, weight int, f core.CreateFunc) {
	factory.constructors[actorType] = &core.ActorConstructor{
		RegisteraionType: regType,
		Weight:           weight,
		Constructor:      f,
	}
}

func (factory *MockActorFactory) Get(actorType string) *core.ActorConstructor {
	if _, ok := factory.constructors[actorType]; ok {
		return factory.constructors[actorType]
	}

	return nil
}
