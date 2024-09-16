# braid
> 

[![image.png](https://i.postimg.cc/1ztqkfhZ/image.png)](https://postimg.cc/K16jLc89)

[![Go Report Card](https://goreportcard.com/badge/github.com/pojol/braid?style=flat-square)](https://goreportcard.com/report/github.com/pojol/braid)


### [Demo](https://github.com/pojol/braid-demo)

### register event
```go
actor.RegisterEvent("10001", func(e *proto.UserEntity) *actor.DefaultChain {
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
            middleware.TryUpdateUserEntity(),
        },
    }
})
```

### state machine
> Both the handler in the timer and the chain handler in the event run in the same goroutine.
```go
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
system.Pub(ctx, router.Target{ID: "mock_actor_1", Ty: "mock_actor", Ev: "ps_mock_test"}, nil)
```

---

### benchmark