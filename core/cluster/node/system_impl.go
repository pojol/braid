package node

import (
	"context"
	fmt "fmt"
	"sync"

	"github.com/pojol/braid/core"
	"github.com/pojol/braid/core/addressbook"
	"github.com/pojol/braid/def"
	"github.com/pojol/braid/lib/grpc"
	"github.com/pojol/braid/lib/log"
	"github.com/pojol/braid/lib/pubsub"
	"github.com/pojol/braid/lib/span"
	"github.com/pojol/braid/router"
)

type NormalSystem struct {
	addressbook *addressbook.AddressBook
	actoridmap  map[string]core.IActor
	client      *grpc.Client
	ps          *pubsub.Pubsub
	acceptor    *Acceptor
	p           core.SystemParm
	loader      core.IActorLoader
	factory     core.IActorFactory

	sync.RWMutex
}

func buildSystemWithOption(nodid string, loader core.IActorLoader, factory core.IActorFactory, opts ...core.SystemOption) core.ISystem {
	var err error

	p := core.SystemParm{
		Ip:     "127.0.0.1",
		NodeID: nodid,
	}
	for _, opt := range opts {
		opt(&p)
	}

	sys := &NormalSystem{
		actoridmap: make(map[string]core.IActor),
	}

	// init grpc client
	sys.client = grpc.BuildClientWithOption()
	sys.loader = loader

	sys.ps = pubsub.BuildWithOption()

	sys.addressbook = addressbook.New(core.AddressInfo{
		Node: p.NodeID,
		Ip:   p.Ip,
		Port: p.Port,
	})
	sys.p = p

	if p.Port != 0 {
		sys.acceptor, err = NewAcceptor(sys, p.Port)
		if err != nil {
			panic(fmt.Errorf("[system] new acceptor err %v", err.Error()))
		}
	}

	return sys
}

func (sys *NormalSystem) Update() {
	if sys.p.Port != 0 && sys.acceptor != nil {
		sys.acceptor.Update()
	}
}

func (sys *NormalSystem) Loader(ty string) core.IActorBuilder {
	return sys.loader.Builder(ty, sys)
}

func (sys *NormalSystem) AddressBook() core.IAddressBook {
	return sys.addressbook
}

func (sys *NormalSystem) Register(builder core.IActorBuilder) (core.IActor, error) {

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
					return nil, fmt.Errorf("[barid.system] register unique type actor %v in %v", builder.GetType(), sys.p.NodeID)
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
	} else {
		panic(fmt.Errorf("[braid.system] actor %v register err, constructor is nil", builder.GetType()))
	}

	sys.Lock()
	sys.actoridmap[builder.GetID()] = actor
	sys.Unlock()

	log.InfoF("[braid.system] node %v register %v %v succ", sys.addressbook.NodeID, builder.GetType(), builder.GetID())
	return actor, nil
}

func (sys *NormalSystem) Unregister(id string) error {
	// First, check if the actor exists and get it
	sys.RLock()
	actor, exists := sys.actoridmap[id]
	sys.RUnlock()

	if !exists {
		return fmt.Errorf("actor %s not found", id)
	}

	// Call Exit on the actor
	actor.Exit()

	// Remove the actor from the map
	sys.Lock()
	delete(sys.actoridmap, id)
	sys.Unlock()

	// Unregister from the address book
	err := sys.addressbook.Unregister(context.TODO(), id, sys.factory.Get(actor.Type()).Weight)
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

func (sys *NormalSystem) Call(tar router.Target, msg *router.MsgWrapper) error {
	// Set message header information
	msg.Req.Header.Event = tar.Ev
	msg.Req.Header.TargetActorID = tar.ID
	msg.Req.Header.TargetActorType = tar.Ty

	var info core.AddressInfo
	var actor core.IActor
	var err error

	if sys.p.Tracer != nil {
		span, err := sys.p.Tracer.GetSpan(span.SystemCall)
		if err == nil {
			msg.Ctx = span.Begin(msg.Ctx)
			fmt.Println(msg.Req.Header.PrevActorType, "=>", tar.Ty)

			span.SetTag("actor", tar.Ty)
			span.SetTag("event", tar.Ev)
			span.SetTag("id", tar.ID)

			defer span.End(msg.Ctx)
		}
	}

	switch tar.ID {
	case def.SymbolWildcard:
		info, err = sys.addressbook.GetWildcardActor(msg.Ctx, tar.Ty)
		// Check if the wildcard actor is local
		sys.RLock()
		actor, ok := sys.actoridmap[info.ActorId]
		sys.RUnlock()
		if ok {
			return sys.handleLocalCall(actor, msg)
		}
	case def.SymbolLocalFirst:
		actor, info, err = sys.findLocalOrWildcardActor(msg.Ctx, tar.Ty)
		if err != nil {
			return err
		}
		if actor != nil {
			// Local call
			return sys.handleLocalCall(actor, msg)
		}
	default:
		// First, check if it's a local call
		sys.RLock()
		actorp, ok := sys.actoridmap[tar.ID]
		sys.RUnlock()

		if ok {
			return sys.handleLocalCall(actorp, msg)
		}

		// If not local, get from addressbook
		info, err = sys.addressbook.GetByID(msg.Ctx, tar.ID)
	}

	if err != nil {
		return err
	}

	// At this point, we know it's a remote call
	return sys.handleRemoteCall(msg.Ctx, info, msg)
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

func (sys *NormalSystem) handleLocalCall(actorp core.IActor, msg *router.MsgWrapper) error {
	root := msg.Wg.Count() == 0
	if root {
		msg.Done = make(chan struct{})
		ready := make(chan struct{})
		go func() {
			<-ready // Wait for Received to complete
			msg.Wg.Wait()
			close(msg.Done)
		}()

		if err := actorp.Received(msg); err != nil {
			close(ready) // Ensure the ready channel is closed even in case of an error
			return err
		}
		close(ready) // Notify the goroutine that Received has completed

		select {
		case <-msg.Done:
			return nil
		case <-msg.Ctx.Done():
			return fmt.Errorf("actor %v message %v processing timed out", msg.Req.Header.TargetActorID, msg.Req.Header.Event)
		}
	} else {
		return actorp.Received(msg)
	}
}

func (sys *NormalSystem) handleRemoteCall(ctx context.Context, addrinfo core.AddressInfo, msg *router.MsgWrapper) error {
	res := &router.RouteRes{}
	err := sys.client.CallWait(ctx,
		fmt.Sprintf("%s:%d", addrinfo.Ip, addrinfo.Port),
		"/router.Acceptor/routing",
		&router.RouteReq{Msg: msg.Req},
		res)

	if err != nil {
		return err
	}

	msg.Res = res.Msg
	return nil
}

func (sys *NormalSystem) Send(tar router.Target, msg *router.MsgWrapper) error {
	// Set message header information
	msg.Req.Header.Event = tar.Ev
	msg.Req.Header.TargetActorID = tar.ID
	msg.Req.Header.TargetActorType = tar.Ty

	sys.RLock()
	actor, isLocal := sys.actoridmap[tar.ID]
	sys.RUnlock()

	if isLocal {
		// For local actors, use Received directly
		return actor.Received(msg)
	}

	// For remote actors, get address info
	info, err := sys.addressbook.GetByID(msg.Ctx, tar.ID)
	if err != nil {
		return err
	}

	return sys.client.Call(msg.Ctx,
		fmt.Sprintf("%s:%d", info.Ip, info.Port),
		"/router.Acceptor/routing",
		&router.RouteReq{Msg: msg.Req},
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
	if sys.p.Port != 0 {
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
