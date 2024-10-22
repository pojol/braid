package mockdata

import (
	"github.com/pojol/braid/core"
	"github.com/pojol/braid/test/mockdata/mockactors"
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

	factory.constructors["MockUserActor"] = &core.ActorConstructor{
		ID:                  "MockUserActor",
		Name:                "MockUserActor",
		Weight:              80,
		Constructor:         NewUserActor,
		NodeUnique:          false,
		GlobalQuantityLimit: 10000,
		Dynamic:             true,
		Options:             make(map[string]string),
	}

	factory.constructors["MockClacActor"] = &core.ActorConstructor{
		ID:                  "MockClacActor",
		Name:                "MockClacActor",
		Weight:              20,
		Constructor:         NewClacActor,
		NodeUnique:          false,
		GlobalQuantityLimit: 5,
		Dynamic:             true,
		Options:             make(map[string]string),
	}

	factory.constructors["MockDynamicPicker"] = &core.ActorConstructor{
		ID:                  "MockDynamicPicker",
		Name:                "MockDynamicPicker",
		Weight:              100,
		Constructor:         mockactors.NewDynamicPickerActor,
		NodeUnique:          true,
		GlobalQuantityLimit: 10,
		Options:             make(map[string]string),
	}

	factory.constructors["MockDynamicRegister"] = &core.ActorConstructor{
		ID:                  "MockDynamicRegister",
		Name:                "MockDynamicRegister",
		Weight:              100,
		Constructor:         mockactors.NewDynamicRegisterActor,
		NodeUnique:          true,
		GlobalQuantityLimit: 0,
		Options:             make(map[string]string),
	}

	factory.constructors["MockActorControl"] = &core.ActorConstructor{
		ID:                  "MockActorControl",
		Name:                "MockActorControl",
		Weight:              100,
		Constructor:         mockactors.NewControlActor,
		NodeUnique:          true,
		GlobalQuantityLimit: 0,
		Options:             make(map[string]string),
	}

	return factory
}

func (factory *MockActorFactory) Get(actorType string) *core.ActorConstructor {
	if _, ok := factory.constructors[actorType]; ok {
		return factory.constructors[actorType]
	}

	return nil
}

func (factory *MockActorFactory) GetActors() []*core.ActorConstructor {
	actors := []*core.ActorConstructor{}
	for _, v := range factory.constructors {
		actors = append(actors, v)
	}
	return actors
}
