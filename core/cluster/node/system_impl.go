package node

import (
	"context"
	fmt "fmt"
	"sync"

	"github.com/pojol/braid/core"
	"github.com/pojol/braid/core/actor"
	"github.com/pojol/braid/core/addressbook"
	"github.com/pojol/braid/def"
	"github.com/pojol/braid/lib/grpc"
	"github.com/pojol/braid/lib/log"
	"github.com/pojol/braid/lib/pubsub"
	"github.com/pojol/braid/router"
)

type NormalSystem struct {
	addressbook *addressbook.AddressBook
	actoridmap  map[string]core.IActor
	client      *grpc.Client
	ps          *pubsub.Pubsub

	p      SystemParm
	loader core.IActorLoader

	sync.RWMutex
}

func BuildSystemWithOption(nodid string, factory core.IActorFactory, opts ...SystemOption) core.ISystem {

	p := SystemParm{
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
	sys.loader = actor.BuildDefaultActorLoader(sys, factory)

	sys.ps = pubsub.BuildWithOption()

	sys.addressbook = addressbook.New(core.AddressInfo{
		Node: p.NodeID,
		Ip:   p.Ip,
		Port: p.Port,
	})
	sys.p = p

	if p.Port != 0 {
		acceptorInit(sys, p.Port)
	}

	return sys
}

func (sys *NormalSystem) Update() {
	if sys.p.Port != 0 {
		acceptorUpdate()
	}
}

func (sys *NormalSystem) Loader() core.IActorLoader {
	return sys.loader
}

func (sys *NormalSystem) AddressBook() core.IAddressBook {
	return sys.addressbook
}

func (sys *NormalSystem) Register(builder *core.ActorLoaderBuilder) (core.IActor, error) {

	if builder.ID == "" || builder.ActorTy == "" {
		return nil, def.ErrSystemParm()
	}

	sys.Lock()
	if _, ok := sys.actoridmap[builder.ID]; ok {
		sys.Unlock()
		return nil, def.ErrSystemRepeatRegistActor(builder.ActorTy, builder.ID)
	}
	sys.Unlock()

	if builder.GlobalQuantityLimit != 0 {

		// 检查当前节点是否已经存在
		if builder.ActorConstructor.NodeUnique {

		}

		// 检查注册数是否已经超出限制
	}

	// Register first, then build
	err := sys.addressbook.Register(context.TODO(), builder.ActorTy, builder.ID)
	if err != nil {
		return nil, err
	}

	// Instantiate actor
	actor := builder.Constructor(builder)

	sys.Lock()
	sys.actoridmap[builder.ID] = actor
	sys.Unlock()

	log.Info("[braid.system] node %v register %v succ, cur weight %v", sys.addressbook.NodeID, builder.ActorTy, 0)
	return actor, nil
}

func (sys *NormalSystem) Actors() []core.IActor {
	actors := []core.IActor{}
	for _, v := range sys.actoridmap {
		actors = append(actors, v)
	}
	return actors
}

func (sys *NormalSystem) Call(ctx context.Context, tar router.Target, msg *router.MsgWrapper) error {
	// Set message header information
	msg.Req.Header.Event = tar.Ev
	msg.Req.Header.TargetActorID = tar.ID
	msg.Req.Header.TargetActorType = tar.Ty

	var info core.AddressInfo
	var actor core.IActor
	var err error

	switch tar.ID {
	case def.SymbolWildcard:
		info, err = sys.addressbook.GetWildcardActor(ctx, tar.Ty)
		// Check if the wildcard actor is local
		sys.RLock()
		actor, ok := sys.actoridmap[info.ActorId]
		sys.RUnlock()
		if ok {
			return sys.handleLocalCall(ctx, actor, msg)
		}
	case def.SymbolLocalFirst:
		actor, info, err = sys.findLocalOrWildcardActor(ctx, tar.Ty)
		if err != nil {
			return err
		}
		if actor != nil {
			// Local call
			return sys.handleLocalCall(ctx, actor, msg)
		}
	default:
		// First, check if it's a local call
		sys.RLock()
		actorp, ok := sys.actoridmap[tar.ID]
		sys.RUnlock()

		if ok {
			return sys.handleLocalCall(ctx, actorp, msg)
		}

		// If not local, get from addressbook
		info, err = sys.addressbook.GetByID(ctx, tar.ID)
	}

	if err != nil {
		return err
	}

	// At this point, we know it's a remote call
	return sys.handleRemoteCall(ctx, info, msg)
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

func (sys *NormalSystem) handleLocalCall(ctx context.Context, actorp core.IActor, msg *router.MsgWrapper) error {
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
		case <-ctx.Done():
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

func (sys *NormalSystem) Send(ctx context.Context, tar router.Target, msg *router.MsgWrapper) error {
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
	info, err := sys.addressbook.GetByID(ctx, tar.ID)
	if err != nil {
		return err
	}

	return sys.client.Call(ctx,
		fmt.Sprintf("%s:%d", info.Ip, info.Port),
		"/router.Acceptor/routing",
		&router.RouteReq{Msg: msg.Req},
		nil) // We don't need the response for Send
}

func (sys *NormalSystem) Pub(ctx context.Context, topic string, msg *router.Message) error {

	sys.ps.GetTopic(topic).Pub(ctx, msg)

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

	return nil, def.ErrSystemCantFindLocalActor(id)
}

func (sys *NormalSystem) Exit() {
	if sys.p.Port != 0 {
		acceptorExit()
	}
}
