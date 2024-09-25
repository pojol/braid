# braid
> 

[![image.png](https://i.postimg.cc/3xNDLTwR/image.png)](https://postimg.cc/ts0TT8WQ)

[![Go Report Card](https://goreportcard.com/badge/github.com/pojol/braid?style=flat-square)](https://goreportcard.com/report/github.com/pojol/braid)
[![Demo](https://img.shields.io/badge/demo-braid--demo-brightgreen?style=flat-square)](https://github.com/pojol/braid-demo)
[![Gitter](https://img.shields.io/gitter/room/braid/community?color=blue?style=flat-square)](https://app.gitter.im/#/room/#braid:gitter.im)


### Sample

1. 注册 actor
```go
// factory  e.g. test/mockdata/actor_factory
factory.bind("MockClacActor", 
    core.ActorRegisteraionType_DynamicRandom,  // 动态注册 actor
    20,             // actor 的权重
    50000,          // actor 在集群中的构建数量上限
    NewClacActor,   // actor 的构造函数
)
```

2. 构建 actor
```go

// 构建一个 ActorDynamicRegister 类型的 actor 到本节点中
sys.Loader().Builder(def.ActorDynamicRegister).WithID("nodeid-register").RegisterLocally()

// 或通过 dynamic 的方式将 MockClacActor 类型的 actor 注册到集群（通过负载均衡
sys.Loader().Builder("MockClacActor").WithID("001").RegisterDynamically()
```

3. 为 actor 绑定实现逻辑
```go

// 绑定消息处理
clacActor.RegisterEvent("ev_clac", func(actorCtx context.Context) *actor.DefaultChain {
    
    // 使用中间件
    unpackcfg := &middleware.MsgUnpackCfg[proto.xxx]{}
    sys := core.GetSystem(actorCtx)

    return &actor.DefaultChain{
        Before: []Base.MiddlewareHandler{
            middleware.MsgUnpack(unpackcfg),
        },
        Handler: func(ctx context.Context, msg *router.MsgWrapper) error {

            realmsg, ok := unpackcfg.Msg.(*proto.xxx)
            // todo ...

            // 向下传递消息
            sys.Call(...)

            return nil
        }
    }
})

// 绑定定时处理函数
clacActor.RegisterTimer(0, 1000, func(actorCtx context.Context) error {

    state := core.GetState(actorCtx).(*xxxState)

    if state.State == Init {
        // todo & state transitions
        state.State = Running
    } else if state.State == Running {

    }

    return nil
})

// Define a message with topic events.EvChatMessageStore and channel a.Id (self)
// func is the callback for successful subscription, registering a handler function for
//   messages returned to the actor
// WithTTL sets the expiration time for this topic to 30 days
err := a.SubscriptionEvent(events.EvChatMessageStore, a.Id, func() {
    a.RegisterEvent(events.EvChatMessageStore, events.MakeChatStoreMessage)
}, pubsub.WithTTL(time.Hour*24*30))
if err != nil {
    log.Warn("actor %v ty %v subscription event %v err %v", a.Id, a.Ty, ev, err.Error())
}
```


---

### benchmark