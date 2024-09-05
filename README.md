# braid

[![image.png](https://i.postimg.cc/MGs0hDTP/image.png)](https://postimg.cc/jW7JfyyP)

### register event
```go
actor.RegisterEvent("10001", func(e *proto.UserEntity) *workerthread.DefaultChain {
    unpackcfg := &middleware.MsgUnpackCfg[proto.GetUserInfoReq]{}

    return &workerthread.DefaultChain{
        Before: []workerthread.MiddlewareHandler{
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

### register timer

### state machine


---

### benchmark