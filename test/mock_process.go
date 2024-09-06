package test

import (
	"fmt"
	"os"
	"os/signal"
	"syscall"

	"github.com/pojol/braid/core/cluster/node"
	"github.com/pojol/braid/core/workerthread"
)

type ProcessNode struct {
	P   node.Parm
	Sys workerthread.ISystem
}

func New() *ProcessNode {
	return &ProcessNode{}
}

func (pn *ProcessNode) ID() string {
	return pn.P.ID
}

func (pn *ProcessNode) System() workerthread.ISystem {
	return pn.Sys
}

func (pn *ProcessNode) Init(opts ...node.Option) error {

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
