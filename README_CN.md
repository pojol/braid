# Braid 一个轻量级的 Actor 游戏开发框架
> Braid 是一个以 Actor 模型为核心驱动的创新型无服务器游戏框架。它通过统一的寻址系统实现自动负载管理，使开发者能够专注于设计和实现，而无需关心复杂的分布式系统组件。

[![Go Report Card](https://goreportcard.com/badge/github.com/pojol/braid?style=flat-square)](https://goreportcard.com/report/github.com/pojol/braid)
[![Demo](https://img.shields.io/badge/demo-braid--demo-brightgreen?style=flat-square)](https://github.com/pojol/braid-demo)
[![Documentation](https://img.shields.io/badge/Documentation-Available-brightgreen)](https://pojol.github.io/braid/#/)
[![Discord](https://img.shields.io/discord/1210543471593791488?color=7289da&label=Discord&logo=discord&style=flat-square)](https://discord.gg/yXJgTrkWxT)

[![image.png](https://i.postimg.cc/BbvzLhfN/image.png)](https://postimg.cc/Vr3g2W6b)

### 特性
* 以 Actor 为中心：框架本质上是 Actor 的集合，简化了分布式逻辑。
* 自动负载均衡：通过寻址系统实现自动资源分配。
* 专注开发：无需考虑服务或集群等底层架构，专注于游戏逻辑。

### 1. 快速开始
> 使用 braid-cli 工具安装脚手架项目

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
$ go run main.go
```

### 2. 创建新的 actor 并将其加载到集群中
> 编写 node.yaml 将 actor 模板注册到节点（容器）中

```yaml
actors:
- name: "USER"
    id : "user"
    unique: false
    weight: 100
    limit: 10000
```
> 将 actor 的构建函数绑定到 actor 工厂中

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

### 3. 实现 actor 的逻辑
> 注意：在 actor 中注册的所有处理函数（事件、定时器）都是同步处理的，用户无需关心集群内部的异步逻辑。

> 绑定事件函数具柄
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
> 绑定 timer handler
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
> 订阅消息（mq
```go
user.SubscriptionEvent("offline_messages", a.Id, func() {

    // After successful subscription, bind a handler function for the message
    a.RegisterEvent(events.EvChatMessageStore, events.MakeChatStoreMessage)
    
}, pubsub.WithTTL(time.Hour*24*30))
```


---

### benchmark