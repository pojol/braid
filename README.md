# Braid
> A high-performance distributed framework powered by Actor model, designed for building scalable microservices and real-time applications with ease.

[![Go Report Card](https://goreportcard.com/badge/github.com/pojol/braid?style=flat-square)](https://goreportcard.com/report/github.com/pojol/braid)
[![Documentation](https://img.shields.io/badge/Documentation-Available-brightgreen)](https://pojol.github.io/braid/#/)
<!--
[![Discord](https://img.shields.io/discord/1210543471593791488?color=7289da&label=Discord&logo=discord&style=flat-square)](https://discord.gg/yXJgTrkWxT)
-->

[![image.png](https://i.postimg.cc/pr9vjVDm/image.png)](https://postimg.cc/T5XBM6Nx)

[中文](https://github.com/pojol/braid/blob/master/README_CN.md)

## Core Features

- **Lightweight Actor Model**: Efficient Actor system based on Go goroutines, where each Actor is an independent computation unit
- **Flexible Message Routing**: Supports point-to-point communication, broadcasting, and wildcard routing
- **Distributed Addressing**: Built-in distributed address book with dynamic service discovery and load balancing
- **High-Performance Communication**: Efficient inter-node communication based on gRPC
- **Observability**: Built-in tracing and monitoring support
- **Fault Tolerance**: Built-in fault recovery and error handling mechanisms
- **Pub/Sub**: Topic-based message publishing and subscription support

## Use Cases

- **Game Servers**: Ideal for handling large numbers of concurrent users and real-time communication
- **IoT Applications**: Managing large-scale device connections and message routing
- **Microservices Architecture**: Building scalable distributed service systems
- **Real-time Data Processing**: Handling high-concurrency data and event streams
- **Distributed Computing**: Supporting complex distributed computation tasks

## Advantages

1. **Easy to Use**: Provides intuitive APIs, reducing the complexity of distributed system development
2. **High Performance**: Delivers exceptional performance leveraging Go's concurrency features
3. **Scalability**: Supports horizontal scaling to easily handle business growth
4. **Reliability**: Built-in recovery mechanisms enhance system stability
5. **Development Efficiency**: Offers a complete toolkit to accelerate development cycles

</br>

### 1. Quick Start
> Install and set up a minimal working game server using the braid-cli tool

```shell
# 1. Install CLI Tool
$ go install github.com/pojol/braid-cli@latest

# 2. Generate a New Project 
$ braid-cli new "you-project-name" v0.1.8

# 3. Creating .go Files from Actor Template Configurations
$ cd you-project-name/template
$ go generate

# 4. Navigate to the services directory, then try to build and run the demo
$ cd you-project-name/node
$ go run main.go
```

### 2. Implement logic for the actor

```golang
user.OnEvent("xx_event", func(ctx core.ActorContext) *actor.DefaultChain {
    // use unpack middleware
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

### 3. Message sending

```golang
m := msg.NewBuilder(context.TODO())
m.WithReqCustomFields(fields.RoomID(b.RoomID))
ctx.Call(b.ID, template.ACTOR_USER, constant.Ev_UpdateUserInfo, m.Build())
```


### 4. Built-in Support for Jaeger Distributed Tracing
[![image.png](https://i.postimg.cc/wTVhQhyM/image.png)](https://postimg.cc/XprGVBg6)

<div style="display: flex; align-items: center; margin: 1em 0;">
  <div style="flex-grow: 1; height: 1px; background-color: #ccc;"></div>
  <div style="margin: 0 10px; font-weight: bold; color: #666;">Testing Robot</div>
  <div style="flex-grow: 1; height: 1px; background-color: #ccc;"></div>
</div>

### 5. Testing with Bot Framework
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
```shell
goos: darwin
goarch: amd64
cpu: VirtualApple @ 2.50GHz
```
| Test Item | Node Count | Performance |
|-----------|------------|-------------|
| [dynamic-picker](https://github.com/pojol/braid/blob/master/tests/addressbook_test.go) | 10 | 500 actors/s |
| [call](https://github.com/pojol/braid/blob/master/tests/call_benchmark_test.go) | 2 (a1 -> a2 -> b1) | 14000 calls/s |
