package node

import (
	"context"
	"os"
	"os/signal"
	"sync"
	"syscall"
	"time"

	"github.com/pojol/braid/core"
	"github.com/pojol/braid/lib/log"
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

	return nil
}

func (pn *process) WaitClose() {
	ch := make(chan os.Signal, 1)
	signal.Notify(ch, syscall.SIGTERM, syscall.SIGINT, syscall.SIGQUIT)
	s := <-ch
	log.InfoF("Received signal %v, initiating graceful shutdown...", s)

	// Create a WaitGroup and a context with timeout
	var wg sync.WaitGroup
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	// Call the system's shutdown method with the WaitGroup
	pn.sys.Exit(&wg)

	// Create a channel to signal when all actors have finished
	done := make(chan struct{})

	// Wait for all actors to finish their cleanup in a goroutine
	go func() {
		wg.Wait()
		close(done)
	}()

	// Wait for either all actors to finish or the timeout to occur
	select {
	case <-done:
		log.InfoF("All actors have shut down gracefully. Exiting process.")
	case <-ctx.Done():
		log.InfoF("Shutdown timed out after 30 seconds. Force exiting.")
	}

	// Perform any final cleanup if necessary
	// ...

	log.InfoF("Process exited.")
}
