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

var pcs *process

func BuildProcessWithOption(opts ...Option) INode {

	p := Parm{
		Ip: "127.0.0.1",
	}

	for _, opt := range opts {
		opt(&p)
	}

	pcs = &process{
		p: p,
	}

	return pcs
}

func Get() INode {
	return pcs
}

func (pn *process) ID() string {
	return pn.p.ID
}

func (pn *process) System() workerthread.ISystem {
	return pn.p.Sys
}

func (pn *process) Init(opts ...Option) error {

	for _, a := range pn.p.Sys.Actors() {
		a.Init()
	}

	return nil
}

func (pn *process) Update() {

	pn.p.Sys.Update()

	for _, a := range pn.p.Sys.Actors() {
		go a.Update()
	}
}

func (pn *process) WaitClose() {
	ch := make(chan os.Signal, 1)
	signal.Notify(ch, syscall.SIGTERM, syscall.SIGINT, syscall.SIGQUIT)
	s := <-ch
	fmt.Printf("signal %v\n", s)
}
