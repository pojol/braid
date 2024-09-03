package node

import (
	"fmt"
	"os"
	"os/signal"
	"syscall"

	"github.com/pojol/braid/core/workerthread"
)

type process struct {
	p Parm
}

func NewProcess() INode {
	return &process{}
}

func (pn *process) ID() string {
	return pn.p.ID
}

func (pn *process) Name() string {
	return pn.p.Name
}

func (pn *process) Init(opts ...Option) error {

	for _, a := range workerthread.Actors() {
		a.Init()
	}

	return nil
}

func (pn *process) Update(actors ...workerthread.IActor) {
	for _, a := range actors {
		go a.Update()
	}
}

func (pn *process) WaitClose() {
	ch := make(chan os.Signal, 1)
	signal.Notify(ch, syscall.SIGTERM, syscall.SIGINT, syscall.SIGQUIT)
	s := <-ch
	fmt.Printf("signal %v\n", s)
}
