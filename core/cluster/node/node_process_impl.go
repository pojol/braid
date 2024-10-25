package node

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"sync"
	"syscall"

	"github.com/pojol/braid/core"
)

type process struct {
	p   core.NodeParm
	sys core.ISystem
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
		sys: buildSystemWithOption(p.ID, p.Ip, p.Port, p.Loader, p.Factory, p.Tracer),
		p:   p,
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
	return pn.sys
}

func (pn *process) Init(opts ...core.NodeOption) error {

	pn.p.Loader.AssignToNode(pn)

	for _, a := range pn.sys.Actors() {
		a.Init(context.TODO())
	}

	return nil
}

func (pn *process) Update() {

	pn.sys.Update()

	for _, a := range pn.sys.Actors() {
		go a.Update()
	}
}

func (pn *process) WaitClose() {
	ch := make(chan os.Signal, 1)
	signal.Notify(ch, syscall.SIGTERM, syscall.SIGINT, syscall.SIGQUIT)
	s := <-ch
	fmt.Printf("Received signal %v, initiating graceful shutdown...\n", s)

	// Create a WaitGroup
	var wg sync.WaitGroup

	// Call the system's shutdown method with the WaitGroup
	pn.sys.Exit(&wg)

	// Wait for all actors to finish their cleanup
	wg.Wait()

	fmt.Println("All actors have shut down. Exiting process.")
}
