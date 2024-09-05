package nodeprocess

import (
	"fmt"
	"os"
	"os/signal"
	"syscall"

	"github.com/pojol/braid/core/cluster/node"
	"github.com/pojol/braid/core/workerthread"
)

type ProcessNode struct {
	p   node.Parm
	sys workerthread.ISystem
}

func New() *ProcessNode {
	return &ProcessNode{}
}

func (pn *ProcessNode) ID() string {
	return pn.p.ID
}

func (pn *ProcessNode) System() workerthread.ISystem {
	return pn.sys
}

func (pn *ProcessNode) Init(opts ...node.Option) error {

	for _, a := range pn.sys.Actors() {
		a.Init()
	}

	return nil
}

func (pn *ProcessNode) Update() {
	for _, a := range pn.sys.Actors() {
		go a.Update()
	}
}

func (pn *ProcessNode) WaitClose() {
	ch := make(chan os.Signal, 1)
	signal.Notify(ch, syscall.SIGTERM, syscall.SIGINT, syscall.SIGQUIT)
	s := <-ch
	fmt.Printf("signal %v\n", s)
}
