package actor

import (
	"github.com/pojol/braid/core"
)

type ActorGenerationMode int

const (
	LocalGeneration ActorGenerationMode = iota
	BalancedGeneration
)

type CreateActorParm struct {
	ID             string
	ActorTy        string
	Options        map[string]interface{}
	GenerationMode ActorGenerationMode
}

// ActorLoaderBuilder used to build ActorLoader
type ActorLoaderBuilder struct {
	CreateActorParm
	core.ISystem
	core.ActorConstructor
	core.IActorLoader
}

func (p *ActorLoaderBuilder) WithID(id string) core.IActorBuilder {
	if id == "" {
		panic("[braid.actor] id is empty")
	}
	p.ID = id
	return p
}

func (p *ActorLoaderBuilder) WithType(ty string) core.IActorBuilder {
	p.ActorTy = ty
	return p
}

func (p *ActorLoaderBuilder) WithOpt(key string, value interface{}) core.IActorBuilder {
	p.Options[key] = value
	return p
}

func (p *ActorLoaderBuilder) WithPicker() core.IActorBuilder {
	p.GenerationMode = BalancedGeneration
	return p
}

func (p *ActorLoaderBuilder) Build() (core.IActor, error) {
	var err error
	if p.GenerationMode == LocalGeneration {
		return p.ISystem.Register(p)
	} else if p.GenerationMode == BalancedGeneration {
		err = p.IActorLoader.Pick(p) // Note: This method is asynchronous
	}
	return nil, err
}

func (p *ActorLoaderBuilder) GetID() string {
	return p.ID
}

func (p *ActorLoaderBuilder) GetType() string {
	return p.ActorTy
}

func (p *ActorLoaderBuilder) GetGlobalQuantityLimit() int {
	return p.GlobalQuantityLimit
}

func (p *ActorLoaderBuilder) GetNodeUnique() bool {
	return p.NodeUnique
}

func (p *ActorLoaderBuilder) GetOptions() map[string]interface{} {
	return p.Options
}

func (p *ActorLoaderBuilder) GetOpt(key string) interface{} {
	return p.Options[key]
}

func (p *ActorLoaderBuilder) GetSystem() core.ISystem {
	return p.ISystem
}

func (p *ActorLoaderBuilder) GetLoader() core.IActorLoader {
	return p.IActorLoader
}

func (p *ActorLoaderBuilder) GetConstructor() core.CreateFunc {
	return p.Constructor
}
