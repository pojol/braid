package node

import (
	"fmt"
	"os"
	"os/signal"
	"syscall"

	"github.com/pojol/braid/core"
)

type process struct {
	p core.NodeParm
}

var pcs *process

func BuildProcessWithOption(opts ...core.NodeOption) core.INode {

	p := core.NodeParm{
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

func Get() core.INode {
	return pcs
}

func (pn *process) ID() string {
	return pn.p.ID
}

func (pn *process) System() core.ISystem {
	return pn.p.Sys
}

func (pn *process) Init(opts ...core.NodeOption) error {

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
