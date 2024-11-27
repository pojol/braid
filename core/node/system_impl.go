package node

import (
	"context"
	"errors"
	fmt "fmt"
	"sync"

	"github.com/pojol/braid/core"
	"github.com/pojol/braid/core/addressbook"
	"github.com/pojol/braid/def"
	"github.com/pojol/braid/lib/grpc"
	"github.com/pojol/braid/lib/log"
	"github.com/pojol/braid/lib/pubsub"
	"github.com/pojol/braid/lib/span"
	"github.com/pojol/braid/lib/tracer"
	"github.com/pojol/braid/router"
	"github.com/pojol/braid/router/msg"
)

type NormalSystem struct {
	addressbook *addressbook.AddressBook
	actoridmap  map[string]core.IActor
	client      *grpc.Client
	ps          *pubsub.Pubsub
	acceptor    *Acceptor
	loader      core.IActorLoader
	factory     core.IActorFactory

	nodeID   string
	nodeIP   string
	nodePort int

	trac tracer.ITracer

	sync.RWMutex
}

var ErrSelfCall = errors.New("cannot call self node through RPC")

func buildSystemWithOption(nodId, nodeIp string, nodePort int, loader core.IActorLoader, factory core.IActorFactory, trac tracer.ITracer) core.ISystem {
	var err error

	sys := &NormalSystem{
		actoridmap: make(map[string]core.IActor),
		nodeID:     nodId,
		nodeIP:     nodeIp,
		nodePort:   nodePort,
		trac:       trac,
	}

	if loader == nil || factory == nil {
		panic("[braid.system] loader or factory is nil!")
	}

	// init grpc client
	sys.client = grpc.BuildClientWithOption()
	sys.loader = loader
	sys.factory = factory

	sys.ps = pubsub.BuildWithOption()

	sys.addressbook = addressbook.New(core.AddressInfo{
		Node: sys.nodeID,
		Ip:   sys.nodeIP,
		Port: sys.nodePort,
	})

	if sys.nodePort != 0 {
		sys.acceptor, err = NewAcceptor(sys, sys.nodePort)
		if err != nil {
			panic(fmt.Errorf("[braid.system] new acceptor err %v", err.Error()))
		}

		// run grpc acceptor
		sys.acceptor.server.Run()
	}

	return sys
}

func (sys *NormalSystem) Loader(ty string) core.IActorBuilder {
	return sys.loader.Builder(ty, sys)
}

func (sys *NormalSystem) AddressBook() core.IAddressBook {
	return sys.addressbook
}

func (sys *NormalSystem) Register(ctx context.Context, builder core.IActorBuilder) (core.IActor, error) {

	if builder.GetID() == "" || builder.GetType() == "" {
		return nil, fmt.Errorf("[braid.system] register actor id %v type %v parm err", builder.GetID(), builder.GetType())
	}

	sys.Lock()
	if _, ok := sys.actoridmap[builder.GetID()]; ok {
		sys.Unlock()
		return nil, fmt.Errorf("[braid.system] register actor %v repeat", builder.GetID())
	}
	sys.Unlock()

	if builder.GetGlobalQuantityLimit() != 0 {

		if builder.GetNodeUnique() {
			for _, v := range sys.actoridmap {
				if v.Type() == builder.GetType() {
					return nil, fmt.Errorf("[barid.system] register unique type actor %v in %v", builder.GetType(), sys.nodeID)
				}
			}
		}

		cnt, err := sys.addressbook.GetActorTypeCount(context.TODO(), builder.GetType())
		if err != nil {
			return nil, fmt.Errorf("[barid.system] get type count err %v", err)
		}
		if int(cnt) >= builder.GetGlobalQuantityLimit() {
			return nil, fmt.Errorf("[braid.system] actor %v global quantity limit current count %v", builder.GetType(), cnt)
		}
	}

	// Register first, then build
	err := sys.addressbook.Register(context.TODO(), builder.GetType(), builder.GetID(), builder.GetWeight())
	if err != nil {
		return nil, err
	}

	// Instantiate actor
	var actor core.IActor
	if builder.GetConstructor() != nil {
		actor = builder.GetConstructor()(builder)
		actor.Init(ctx)
	} else {
		panic(fmt.Errorf("[braid.system] actor %v register err, constructor is nil", builder.GetType()))
	}

	sys.Lock()
	sys.actoridmap[builder.GetID()] = actor
	sys.Unlock()

	log.InfoF("[braid.system] node %v register %v %v succ", sys.addressbook.NodeID, builder.GetType(), builder.GetID())
	return actor, nil
}

func (sys *NormalSystem) Unregister(id, ty string) error {
	// First, check if the actor exists and get it
	sys.RLock()
	actor, exists := sys.actoridmap[id]
	sys.RUnlock()

	if exists {
		// Call Exit on the actor
		actor.Exit()

		// Remove the actor from the map
		sys.Lock()
		delete(sys.actoridmap, id)
		sys.Unlock()
	}

	// Unregister from the address book
	ac := sys.factory.Get(ty)
	if ac == nil {
		return fmt.Errorf("[braid.system] unregister actor id %v unknown type %v", id, ty)
	}

	err := sys.addressbook.Unregister(context.TODO(), id, sys.factory.Get(ty).Weight)
	if err != nil {
		// Log the error, but don't return it as the actor has already been removed locally
		log.WarnF("[braid.system] Failed to unregister actor %s from address book: %v", id, err)
	}

	log.InfoF("[braid.system] Actor %s unregistered successfully", id)

	return nil
}

func (sys *NormalSystem) Actors() []core.IActor {
	actors := []core.IActor{}
	for _, v := range sys.actoridmap {
		actors = append(actors, v)
	}
	return actors
}

func (sys *NormalSystem) Call(idOrSymbol, actorType, event string, mw *msg.Wrapper) error {
	// Set message header information
	mw.Req.Header.Event = event
	mw.Req.Header.TargetActorID = idOrSymbol
	mw.Req.Header.TargetActorType = actorType

	var info core.AddressInfo
	var actor core.IActor
	var err error

	if sys.trac != nil {
		span, err := sys.trac.GetSpan(span.SystemCall)
		if err == nil {
			mw.Ctx = span.Begin(mw.Ctx)

			span.SetTag("actor", actorType)
			span.SetTag("event", event)
			span.SetTag("id", idOrSymbol)

			defer span.End(mw.Ctx)
		}
	}

	if idOrSymbol == "" {
		return fmt.Errorf("[braid.system] call unknown target id")
	}

	switch idOrSymbol {
	case def.SymbolWildcard:
		info, err = sys.addressbook.GetWildcardActor(mw.Ctx, actorType)
		// Check if the wildcard actor is local
		sys.RLock()
		actor, ok := sys.actoridmap[info.ActorId]
		sys.RUnlock()
		if ok {
			return sys.localCall(actor, mw)
		}
	case def.SymbolLocalFirst:
		actor, info, err = sys.findLocalOrWildcardActor(mw.Ctx, actorType)
		if err != nil {
			return err
		}
		if actor != nil {
			// Local call
			return sys.localCall(actor, mw)
		}
	default:
		// First, check if it's a local call
		sys.RLock()
		actorp, ok := sys.actoridmap[idOrSymbol]
		sys.RUnlock()

		if ok {
			return sys.localCall(actorp, mw)
		}

		// If not local, get from addressbook
		info, err = sys.addressbook.GetByID(mw.Ctx, idOrSymbol)
	}

	if err != nil {
		return fmt.Errorf("[braid.system] call id %v ty %v err %w", idOrSymbol, actorType, err)
	}

	if info.Ip == sys.nodeIP && info.Port == sys.nodePort {
		log.WarnF("[braid.system] call err actorTy %v actorID %v call ev %v self-call", actorType, idOrSymbol, event)
		return ErrSelfCall
	}

	// At this point, we know it's a remote call
	return sys.handleRemoteCall(mw.Ctx, info, mw)
}

func (sys *NormalSystem) findLocalOrWildcardActor(ctx context.Context, ty string) (core.IActor, core.AddressInfo, error) {
	sys.RLock()
	defer sys.RUnlock()

	for id, actor := range sys.actoridmap {
		if actor.Type() == ty {
			return actor, core.AddressInfo{ActorId: id, ActorTy: ty}, nil
		}
	}

	// If not found locally, use GetWildcardActor to perform a random search across the cluster
	info, err := sys.addressbook.GetWildcardActor(ctx, ty)
	return nil, info, err
}

func (sys *NormalSystem) localCall(actorp core.IActor, mw *msg.Wrapper) error {
	root := mw.Wg.Count() == 0
	if root {
		mw.Done = make(chan struct{})
		ready := make(chan struct{})
		go func() {
			<-ready // Wait for Received to complete
			mw.Wg.Wait()
			close(mw.Done)
		}()

		if err := actorp.Received(mw); err != nil {
			close(ready) // Ensure the ready channel is closed even in case of an error
			return err
		}
		close(ready) // Notify the goroutine that Received has completed

		select {
		case <-mw.Done:
			return nil
		case <-mw.Ctx.Done():
			return fmt.Errorf("[braid.system] actor %v message %v processing timed out", mw.Req.Header.TargetActorID, mw.Req.Header.Event)
		}
	} else {
		return actorp.Received(mw)
	}
}

func (sys *NormalSystem) handleRemoteCall(ctx context.Context, addrinfo core.AddressInfo, mw *msg.Wrapper) error {
	res := &router.RouteRes{}
	err := sys.client.CallWait(ctx,
		fmt.Sprintf("%s:%d", addrinfo.Ip, addrinfo.Port),
		"/router.Acceptor/routing",
		&router.RouteReq{Msg: mw.Req},
		res)

	if err != nil {
		return err
	}

	mw.Res = res.Msg
	return nil
}

func (sys *NormalSystem) Send(idOrSymbol, actorType, event string, mw *msg.Wrapper) error {
	// Set message header information
	mw.Req.Header.Event = event
	mw.Req.Header.TargetActorID = idOrSymbol
	mw.Req.Header.TargetActorType = actorType

	var info core.AddressInfo
	var actor core.IActor
	var err error

	if idOrSymbol == "" {
		return fmt.Errorf("[braid.system] send unknown target id")
	}

	switch idOrSymbol {
	case def.SymbolWildcard:
		info, err = sys.addressbook.GetWildcardActor(mw.Ctx, actorType)
		// Check if the wildcard actor is local
		sys.RLock()
		actor, ok := sys.actoridmap[info.ActorId]
		sys.RUnlock()
		if ok {
			return actor.Received(mw)
		}
	case def.SymbolLocalFirst:
		actor, info, err = sys.findLocalOrWildcardActor(mw.Ctx, actorType)
		if err != nil {
			return err
		}
		if actor != nil {
			return actor.Received(mw)
		}
	default:
		// First, check if it's a local call
		sys.RLock()
		actorp, ok := sys.actoridmap[idOrSymbol]
		sys.RUnlock()

		if ok {
			return actorp.Received(mw)
		}

		// If not local, get from addressbook
		info, err = sys.addressbook.GetByID(mw.Ctx, idOrSymbol)
	}

	if err != nil {
		return fmt.Errorf("[braid.system] send id %v ty %v err %w", idOrSymbol, actorType, err)
	}

	if info.Ip == sys.nodeIP && info.Port == sys.nodePort {
		log.WarnF("[braid.system] send err actorTy %v actorID %v call ev %v self-call", actorType, idOrSymbol, event)
		return ErrSelfCall
	}

	return sys.handleRemoteSend(info, mw)
}

func (sys *NormalSystem) handleRemoteSend(info core.AddressInfo, mw *msg.Wrapper) error {
	return sys.client.Call(mw.Ctx,
		fmt.Sprintf("%s:%d", info.Ip, info.Port),
		"/router.Acceptor/routing",
		&router.RouteReq{Msg: mw.Req},
		nil) // We don't need the response for Send
}

func (sys *NormalSystem) Pub(topic string, msg *router.Message) error {

	sys.ps.GetTopic(topic).Pub(context.TODO(), msg)

	return nil
}

func (sys *NormalSystem) Sub(topic string, channel string, opts ...pubsub.TopicOption) (*pubsub.Channel, error) {
	return sys.ps.GetOrCreateTopic(topic).Sub(context.TODO(), channel)
}

func (sys *NormalSystem) FindActor(ctx context.Context, id string) (core.IActor, error) {
	sys.RLock()
	defer sys.RUnlock()

	if _, ok := sys.actoridmap[id]; ok {
		actorp := sys.actoridmap[id]
		return actorp, nil
	}

	return nil, fmt.Errorf("[braid.system] find actor %v err", id)
}

func (sys *NormalSystem) Exit(wait *sync.WaitGroup) {
	if sys.nodePort != 0 {
		wait.Add(1)
		if sys.acceptor != nil {
			sys.acceptor.Exit()
		}
		wait.Done()
	}

	for _, actor := range sys.actoridmap {
		wait.Add(1)

		go func(a core.IActor) {
			defer wait.Done()
			a.Exit()
		}(actor)
	}

	err := sys.addressbook.Clear(context.TODO())
	if err != nil {
		log.WarnF("[braid.addressbook] clear err %v", err.Error())
	}
}
