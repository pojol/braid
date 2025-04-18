package core

import "context"

type AddressInfo struct {
	ActorId string `json:"actor_id"`
	ActorTy string `json:"actor_ty"`
	Node    string `json:"node"`
	Service string `json:"service"`
	Ip      string `json:"ip"`
	Port    int    `json:"port"`
}

type NodeInfo struct {
	NodeID string
	Ip     string
	Port   int
	Weight int
}

type IAddressBook interface {
	//
	Register(context.Context, string, string, int) error
	Unregister(context.Context, string, int) error

	GetByID(context.Context, string) (AddressInfo, error)
	GetByType(context.Context, string) ([]AddressInfo, error)

	GetLowWeightNodeForActor(ctx context.Context, actorType string) (AddressInfo, error)
	GetActorTypeCount(ctx context.Context, actorType string) (int64, error)

	Clear(context.Context) error
}
