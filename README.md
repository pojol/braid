# Braid: A Lightweight Actor Framework for Game Development
> Braid is an innovative serverless game framework powered by the Actor model. It achieves intelligent load management through a unified addressing system, allowing developers to focus on designing and implementing Actors without the need to concern themselves with complex distributed system components.


[![Go Report Card](https://goreportcard.com/badge/github.com/pojol/braid?style=flat-square)](https://goreportcard.com/report/github.com/pojol/braid)
[![Documentation](https://img.shields.io/badge/Documentation-Available-brightgreen)](https://pojol.github.io/braid/#/)
<!--
[![Discord](https://img.shields.io/discord/1210543471593791488?color=7289da&label=Discord&logo=discord&style=flat-square)](https://discord.gg/yXJgTrkWxT)
-->

[![image.png](https://i.postimg.cc/pr9vjVDm/image.png)](https://postimg.cc/T5XBM6Nx)

[中文](https://github.com/pojol/braid/blob/master/README_CN.md)

### Features
* Actor-Centric: Built around a pure Actor model, simplifying distributed logic.
* Automatic Load Balancing: Intelligent resource allocation through the addressing system.
* Development Focus: No need to consider underlying architecture like services or clusters; concentrate on game logic.

### 1. Quick Start
> Install and set up a minimal working game server using the braid-cli tool

```shell
# 1. Install CLI Tool
$ go install github.com/pojol/braid-cli@latest

# 2. Generate a New Project 
$ braid-cli new "you-project-name" v0.1.3

# 3. Creating .go Files from Actor Template Configurations
$ cd you-project-name/template
$ go generate

# 4. Navigate to the services directory, then try to build and run the demo
$ cd you-project-name/node
$ go run main.go
```

### 2. Create a new actor and load it into the cluster
> Register actors to nodes by configuring `node.yaml`

```yaml
actors:
- name: "USER"
    id : "user"
    unique: false
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
    a.state.Load(ctx)   // Load data from cache to local storage
}

// factory.go with node.yaml
case template.USER:
    factory.bind("USER", v.Unique, v.Weight, v.Limit, NewUserActor)
```

### 3. Implement logic for the actor
> Note: All handling functions (events, timers) registered in an actor are processed synchronously. Users do not need to concern themselves with asynchronous logic within the actor.

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
user.SubscriptionEvent(events.EvChatMessageStore, a.Id, func() {

    // After successful subscription, bind a handler function for the message
    a.RegisterEvent(events.EvChatMessageStore, events.MakeChatStoreMessage)
    
}, pubsub.WithTTL(time.Hour*24*30))
```

<div style="display: flex; align-items: center; margin: 1em 0;">
  <div style="flex-grow: 1; height: 1px; background-color: #ccc;"></div>
  <div style="margin: 0 10px; font-weight: bold; color: #666;">Testing Robot</div>
  <div style="flex-grow: 1; height: 1px; background-color: #ccc;"></div>
</div>

### 4. Testing with Bot Framework
> Use the project built with scaffold above

```shell
$ cd you-project-name/testbots

# 1. Launch Bot service
$ go run main.go

# 2. Download gobot editor #latest
https://github.com/pojol/gobot/releases

# 3. Launch Bot editor
$ run gobot_editor_[ver].exe or .dmg

# 4. Go to Bots tab
# 5. Click Load button to load the bot
# 6. Drag the testbot.bh file from the testbots directory to the bots page
# 7. Click bottom-left Create Bot button to create instance
# 8. Click Run to the Next button to execute the bot step by step. Monitor the bot-server interaction in the right preview window
```

[Gobot](https://github.com/pojol/gobot)
[![image.png](https://i.postimg.cc/LX5gbV34/image.png)](https://postimg.cc/xJrdkMZB)

---

### benchmark
