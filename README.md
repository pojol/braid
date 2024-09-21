# braid
> 

[![image.png](https://i.postimg.cc/3xNDLTwR/image.png)](https://postimg.cc/ts0TT8WQ)

[![Go Report Card](https://goreportcard.com/badge/github.com/pojol/braid?style=flat-square)](https://goreportcard.com/report/github.com/pojol/braid)
[![Demo](https://img.shields.io/badge/demo-braid--demo-brightgreen?style=flat-square)](https://github.com/pojol/braid-demo)
[![Gitter](https://img.shields.io/gitter/room/braid/community?color=blue?style=flat-square)](https://app.gitter.im/#/room/#braid:gitter.im)


### Register event
```go

actor.RegisterEvent("10001", func(sys core.Isystem, state *entity.User) *actor.DefaultChain {
    
    // unpack msg middleware
    unpackcfg := &middleware.MsgUnpackCfg[proto.GetUserInfoReq]{}

    return &actor.DefaultChain{
        Before: []Base.MiddlewareHandler{
            middleware.MsgUnpack(unpackcfg),
        },
        Handler: func(ctx context.Context, msg *router.MsgWrapper) error {

            realmsg, ok := unpackcfg.Msg.(*proto.GetUserInfoReq)
            fmt.Println("recv msg GetUserInfoReq", realmsg)

            // todo ...

            return nil
        }
        After: []workerthread.MiddlewareHandler {
            // Check if the entity is dirty, and if so, synchronize it to the cache
            middleware.TryUpdateUserEntity(),
        },
    }
})
```

### Register timer
> Both the handler in the timer and the chain handler in the event run in the same goroutine.
```go

// 0 execute immediately without waiting
// 1000 execute every 1000 milliseconds
actor.RegisterTimer(0, 1000, func(e *proto.ActivityEntity) error {

    if e.State == Init {
        // todo & state transitions
        e.State = Running
    } else if e.State == Running {

    } else if e.State == Closing {

    } else if e.State == Closed {

    }

    return nil
})

```

### Subscription event
```go

// Define a message with topic events.EvChatMessageStore and channel a.Id (self)
// func is the callback for successful subscription, registering a handler function for
//   messages returned to the actor
// WithTTL sets the expiration time for this topic to 30 days
err := a.SubscriptionEvent(events.EvChatMessageStore, a.Id, func() {
    a.RegisterEvent(events.EvChatMessageStore, events.MakeChatStoreMessage(a.Sys, a.state))
}, pubsub.WithTTL(time.Hour*24*30))
if err != nil {
    log.Warn("actor %v ty %v subscription event %v err %v", a.Id, a.Ty, events.EvChatMessageStore, err.Error())
}
```

### Call
* Sync blocking
```go
// Send a mock_test event to actor_1, blocking and waiting
system.Call(ctx, router.Target{ID: "actor_1", Ty: "mock_actor", Ev: "mock_test"}, nil)
```

* Asyn call
```go
// Send a mock_test event to any actor of type mock_actor
//  async call, and continue execution directly
system.Send(ctx, router.Target{ID:def.SymbolWildcard, Ty: "mock_actor",Ev: "mock_test"}, nil)
```

* Pub
```go
// Publish a ps_mock_test event to mock_actor_1
//    which will be stored in the redis stream queue first
//    waiting for ps_mock_test to consume
system.Pub(ctx, "topic", "channel", msg)
```

---

### benchmark