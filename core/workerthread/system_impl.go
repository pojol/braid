package workerthread

import (
	"context"
	fmt "fmt"
	"sync"

	"github.com/pojol/braid/core/addressbook"
	"github.com/pojol/braid/def"
	"github.com/pojol/braid/lib/grpc"
	"github.com/pojol/braid/router"
)

type NormalSystem struct {
	addressbook *addressbook.AddressBook
	actoridmap  map[string]IActor

	p SystemParm

	sync.RWMutex
}

var sys *NormalSystem

func Init(opts ...SystemOption) {

	p := SystemParm{}
	for _, opt := range opts {
		opt(&p)
	}

	// init grpc client
	grpc.BuildClientWithOption()

	sys.addressbook.NodeID = p.NodeID
	sys.addressbook.ServiceName = p.ServiceName
	sys.p = p
}

func Regist(ty string, opts ...CreateActorOption) (IActor, error) {

	createParm := &CreateActorParm{}
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

	sys.addressbook.Regist(ty, createParm.ID)

	return actor, nil
}

func Actors() []IActor {
	actors := []IActor{}
	for _, v := range sys.actoridmap {
		actors = append(actors, v)
	}
	return actors
}

func Call(ctx context.Context, tar router.Target, msg *router.MsgWrapper) error {

	// 设置消息头部信息
	msg.Req.Header.Event = tar.Ev
	msg.Req.Header.TargetActorID = tar.ID
	msg.Req.Header.TargetActorType = tar.Ty

	// 检查是否为本地调用
	sys.RLock()
	actorp, ok := sys.actoridmap[tar.ID]
	sys.RUnlock()

	if ok {
		return handleLocalCall(ctx, actorp, msg)
	}

	// 处理远程调用
	return handleRemoteCall(ctx, tar.ID, msg)
}

func handleLocalCall(ctx context.Context, actorp IActor, msg *router.MsgWrapper) error {
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

func handleRemoteCall(ctx context.Context, targetID string, msg *router.MsgWrapper) error {
	addrinfo, err := sys.addressbook.GetAddrInfo(targetID)
	if err != nil {
		return err
	}

	res := &router.RouteRes{}
	err = grpc.CallWait(ctx,
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

func Send(ctx context.Context, tar router.Target, msg *router.MsgWrapper) error {
	sys.RLock()
	defer sys.RUnlock()

	return nil
}

func Pub(ctx context.Context, tar router.Target, msg *router.MsgWrapper) error {
	return nil
}

func FindActor(ctx context.Context, id string) (IActor, error) {
	sys.RLock()
	defer sys.RUnlock()

	if _, ok := sys.actoridmap[id]; ok {
		actorp := sys.actoridmap[id]
		return actorp, nil
	}

	return nil, def.ErrSystemCantFindLocalActor(id)
}

func init() {
	sys = &NormalSystem{
		actoridmap:  make(map[string]IActor),
		addressbook: &addressbook.AddressBook{},
	}
}
