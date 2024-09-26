package mockdata

import (
	"fmt"
	"os"
	"os/signal"
	"syscall"

	"github.com/pojol/braid/core"
)

type ProcessNode struct {
	P   core.NodeParm
	Sys core.ISystem
}

func New() *ProcessNode {
	return &ProcessNode{}
}

func (pn *ProcessNode) ID() string {
	return pn.P.ID
}

func (pn *ProcessNode) System() core.ISystem {
	return pn.Sys
}

func (pn *ProcessNode) Init(opts ...core.NodeOption) error {

	for _, a := range pn.Sys.Actors() {
		a.Init()
	}

	return nil
}

func (pn *ProcessNode) Update() {
	pn.Sys.Update()

	for _, a := range pn.Sys.Actors() {
		go a.Update()
	}
}

func (pn *ProcessNode) WaitClose() {
	ch := make(chan os.Signal, 1)
	signal.Notify(ch, syscall.SIGTERM, syscall.SIGINT, syscall.SIGQUIT)
	s := <-ch
	fmt.Printf("signal %v\n", s)
}
