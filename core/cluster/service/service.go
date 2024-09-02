package service

type IService interface {
	ID() string   // node id
	Name() string // service name
}
