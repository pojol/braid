package nodeprocess

import (
	"braid/core/actor"
	"braid/core/cluster/node"
	"fmt"
	"os"
	"os/signal"
	"syscall"
)

type ProcessNode struct {
	p node.Parm
}

func New() *ProcessNode {
	return &ProcessNode{}
}

func (pn *ProcessNode) ID() string {
	return pn.p.ID
}

func (pn *ProcessNode) Name() string {
	return pn.p.Name
}

func (pn *ProcessNode) Init(opts ...node.Option) error {

	for _, a := range actor.Actors() {
		a.Init()
	}

	return nil
}

func (pn *ProcessNode) Update() {
	for _, a := range actor.Actors() {
		go a.Update()
	}
}

func (pn *ProcessNode) WaitClose() {
	ch := make(chan os.Signal, 1)
	signal.Notify(ch, syscall.SIGTERM, syscall.SIGINT, syscall.SIGQUIT)
	s := <-ch
	fmt.Printf("signal %v\n", s)
}
