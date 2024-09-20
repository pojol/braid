package node

import (
	"context"
	fmt "fmt"
	"sync"

	"github.com/pojol/braid/core"
	"github.com/pojol/braid/core/addressbook"
	"github.com/pojol/braid/def"
	"github.com/pojol/braid/lib/grpc"
	"github.com/pojol/braid/lib/pubsub"
	"github.com/pojol/braid/router"
)

type NormalSystem struct {
	addressbook *addressbook.AddressBook
	actoridmap  map[string]core.IActor
	client      *grpc.Client
	ps          *pubsub.Pubsub

	p SystemParm

	sync.RWMutex
}

func BuildSystemWithOption(opts ...SystemOption) core.ISystem {

	p := SystemParm{
		Ip: "127.0.0.1",
	}
	for _, opt := range opts {
		opt(&p)
	}

	sys := &NormalSystem{
		actoridmap: make(map[string]core.IActor),
	}

	// init grpc client
	sys.client = grpc.BuildClientWithOption()

	sys.ps = pubsub.BuildWithOption()

	sys.addressbook = addressbook.New(core.AddressInfo{
		Node:    p.NodeID,
		Service: p.ServiceName,
		Ip:      p.Ip,
		Port:    p.Port,
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

func (sys *NormalSystem) Register(ctx context.Context, ty string, opts ...core.CreateActorOption) (core.IActor, error) {

	createParm := &core.CreateActorParm{
		Sys:     sys,
		Options: make(map[string]interface{}),
	}
	for _, opt := range opts {
		opt(createParm)
	}

	if createParm.ID == "" || ty == "" {
		return nil, def.ErrSystemParm()
	}

	// 检查 actor 是否已存在
	sys.Lock()
	if _, ok := sys.actoridmap[createParm.ID]; ok {
		sys.Unlock()
		return nil, def.ErrSystemRepeatRegistActor(ty, createParm.ID)
	}
	sys.Unlock()

	var creator ActorConstructor
	for _, c := range sys.p.Constructors {
		if c.Type == ty {
			creator = c
			break
		}
	}

	if creator.Type != ty {
		return nil, def.ErrSystemCantFindCreateActorStrategy(ty)
	}

	// 创建 actor
	actor := creator.Constructor(createParm)

	// 注册 actor
	sys.Lock()
	sys.actoridmap[createParm.ID] = actor
	sys.Unlock()

	sys.addressbook.Register(ctx, ty, createParm.ID)

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
			// 如果在本地找到了匹配类型的 actor，直接返回
			return actor, core.AddressInfo{ActorId: id, ActorTy: ty}, nil
		}
	}

	// 如果在本地没有找到，使用 GetWildcardActor 进行集群范围的随机搜索
	info, err := sys.addressbook.GetWildcardActor(ctx, ty)
	return nil, info, err
}

func (sys *NormalSystem) handleLocalCall(ctx context.Context, actorp core.IActor, msg *router.MsgWrapper) error {
	root := msg.Wg.Count() == 0
	if root {
		msg.Done = make(chan struct{})
		ready := make(chan struct{})
		go func() {
			<-ready // 等待 Received 执行完毕
			msg.Wg.Wait()
			close(msg.Done)
		}()

		if err := actorp.Received(msg); err != nil {
			close(ready) // 确保在错误情况下也关闭 ready 通道
			return err
		}
		close(ready) // 通知 goroutine Received 已执行完毕

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

	// Check if it's a local actor
	sys.RLock()
	actor, isLocal := sys.actoridmap[tar.ID]
	sys.RUnlock()

	if isLocal {
		// For local actors, use Received directly
		go actor.Received(msg)
		return nil
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
