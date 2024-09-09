package node

import (
	"context"
	fmt "fmt"
	"sync"

	"github.com/pojol/braid/core"
	"github.com/pojol/braid/core/addressbook"
	"github.com/pojol/braid/def"
	"github.com/pojol/braid/lib/grpc"
	"github.com/pojol/braid/router"
)

type NormalSystem struct {
	addressbook *addressbook.AddressBook
	actoridmap  map[string]core.IActor
	client      *grpc.Client

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
		Sys: sys,
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

	// 设置消息头部信息
	msg.Req.Header.Event = tar.Ev
	msg.Req.Header.TargetActorID = tar.ID
	msg.Req.Header.TargetActorType = tar.Ty
	var err error

	info := core.AddressInfo{ActorId: tar.ID, ActorTy: tar.Ty}

	if /*tar.ID == def.SymbolAll || */ tar.ID == def.SymbolWildcard {

		info, err = sys.addressbook.GetWildcardActor(ctx, tar.Ty)
		if err != nil {
			return err
		}
	}

	// 检查是否为本地调用
	sys.RLock()
	actorp, ok := sys.actoridmap[info.ActorId]
	sys.RUnlock()

	if ok {
		return sys.handleLocalCall(ctx, actorp, msg)
	}

	// 处理远程调用
	return sys.handleRemoteCall(ctx, info.ActorId, msg)
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

func (sys *NormalSystem) handleRemoteCall(ctx context.Context, targetID string, msg *router.MsgWrapper) error {
	addrinfo, err := sys.addressbook.GetByID(ctx, targetID)
	if err != nil {
		return err
	}

	res := &router.RouteRes{}
	err = sys.client.CallWait(ctx,
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
	sys.RLock()
	defer sys.RUnlock()

	return nil
}

func (sys *NormalSystem) Pub(ctx context.Context, tar router.Target, msg *router.MsgWrapper) error {
	return nil
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
