# braid
> 

[![image.png](https://i.postimg.cc/3xNDLTwR/image.png)](https://postimg.cc/ts0TT8WQ)

[![Go Report Card](https://goreportcard.com/badge/github.com/pojol/braid?style=flat-square)](https://goreportcard.com/report/github.com/pojol/braid)
[![Demo](https://img.shields.io/badge/demo-braid--demo-brightgreen?style=flat-square)](https://github.com/pojol/braid-demo)
[![Matrix](https://img.shields.io/badge/chat-%23braid%3Amatrix.org-blue)](https://matrix.to/#/#braid-world:matrix.org)

### Sample

1. 注册 actor
```go
// factory  e.g. test/mockdata/actor_factory
factory.bind("MockClacActor", 
    false,          // 是否节点唯一
    20,             // actor 的权重
    50000,          // actor 在集群中的构建数量上限
    NewClacActor,   // actor 的构造函数
)
```

2. 构建 actor
```go

// 注册一个 ActorDynamicRegister 类型的 actor 到本节点中
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

// 监听消息（离线时别人发来的聊天信息）
//  topic: 离线聊天消息
//  channel: actor自身
//  succ: 成功订阅后的回调
//  ttl: 消息在缓存中保存的时间
clacActor.SubscriptionEvent(events.EvChatMessageStore, a.Id, func() {

    // 监听成功后，为消息绑定处理函数
    a.RegisterEvent(events.EvChatMessageStore, events.MakeChatStoreMessage)
    
}, pubsub.WithTTL(time.Hour*24*30))
```


---

### benchmark