# braid
> 

[![Go Report Card](https://goreportcard.com/badge/github.com/pojol/braid?style=flat-square)](https://goreportcard.com/report/github.com/pojol/braid)
[![Demo](https://img.shields.io/badge/demo-braid--demo-brightgreen?style=flat-square)](https://github.com/pojol/braid-demo)
[![Documentation](https://img.shields.io/badge/Documentation-Available-brightgreen)](https://pojol.github.io/braid/#/)
[![Discord](https://img.shields.io/discord/1210543471593791488?color=7289da&label=Discord&logo=discord&style=flat-square)](https://discord.gg/yXJgTrkWxT)

### 1. Quick Start
> Install the scaffold project using the braid-cli tool

```shell
# 1. Install CLI Tool
$ go install github.com/pojol/braid-cli@latest

# 2. Using the CLI to Generate a New Empty Project
$ braid-cli new "you-project-name"

# 3. Creating .go Files from Actor Template Configurations
$ cd you-project-name/template
$ go generate

# 4. Navigate to the services directory, then try to build and run the demo
$ cd you-project-name/services/demo-1
$ go run main.go~
```

### 2. Create a new actor and load it into the cluster
> Write actor_template.yaml to add new actor types and definitions

```yaml
- name: "USER"
unique: false
category: "dynamic"
```
> Write node.yaml to register actor templates to nodes (containers)

```yaml
actors:
- name: "USER"
    options:
        weight: 100
        limit: 10000
```
> Create actor constructors and bind them to the factory

```golang
type userActor struct {
    *actor.Runtime
    state *Entity
}

func NewUserActor(p core.IActorBuilder) core.IActor {
    return &httpAcceptorActor{
        Runtime: &actor.Runtime{Id: p.GetID(), Ty: p.GetType(), Sys: p.GetSystem()},
        state: user.NewEntity(p.GetID())
    }
}

func (a *userActor) Init(ctx context.Context) {
    a.Runtime.Init(ctx)
    a.state.Load(ctx)   // 将数据从cache装载到本地
}

// factory.go with node.yaml
case template.USER:
    factory.bind(v.Name, v.Unique, v.Weight, v.Limit, NewUserActor)
```

### 3. Implement logic for the actor
> Note: All handling functions (events, timers) registered in the actor are processed synchronously. Users do not need to concern themselves with asynchronous logic within the actor.

> Bind event handler
```go
user.RegisterEvent("use_item", func(ctx core.ActorContext) *actor.DefaultChain {
    // use middleware
    unpackcfg := &middleware.MsgUnpackCfg[proto.xxx]{}

    return &actor.DefaultChain{
        Before: []Base.MiddlewareHandler{
            middleware.MsgUnpack(unpackcfg),
        },
        Handler: func(ctx context.Context, msg *router.MsgWrapper) error {

            realmsg, ok := unpackcfg.Msg.(*proto.xxx)
            // todo ...

            return nil
        }
    }
})
```
> Bind timer handler
```go
user.RegisterTimer(0, 1000, func(ctx core.ActorContext) error {

    state := ctx.GetValue(xxxStateKey{}).(*xxxState)

    if state.State == Init {
        // todo & state transitions
        state.State = Running
    } else if state.State == Running {

    }

    return nil
})
```
> Subscribe to messages and bind event handler
```go
user.SubscriptionEvent("offline_messages", a.Id, func() {

    // After successful subscription, bind a handler function for the message
    a.RegisterEvent(events.EvChatMessageStore, events.MakeChatStoreMessage)
    
}, pubsub.WithTTL(time.Hour*24*30))
```


---

### benchmark