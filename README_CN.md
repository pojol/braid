# Braid
> Braid 是一个轻量级的分布式 Actor 框架，专为构建高性能、可扩展的微服务应用而设计。它提供了一种简单而强大的方式来处理分布式系统中的并发和通信问题。

[![Go Report Card](https://goreportcard.com/badge/github.com/pojol/braid?style=flat-square)](https://goreportcard.com/report/github.com/pojol/braid)
[![文档](https://img.shields.io/badge/文档-Available-brightgreen)](https://pojol.github.io/braid/#/)
<!--
[![Discord](https://img.shields.io/discord/1210543471593791488?color=7289da&label=Discord&logo=discord&style=flat-square)](https://discord.gg/yXJgTrkWxT)
-->

[![image.png](https://i.postimg.cc/BbvzLhfN/image.png)](https://postimg.cc/Vr3g2W6b)

## 核心特性

- **轻量级 Actor 模型**: 基于 Go 协程实现的高效 Actor 系统，每个 Actor 都是独立的计算单元
- **灵活的消息路由**: 支持点对点通信、广播和通配符路由
- **分布式寻址**: 内置分布式地址簿，支持动态服务发现和负载均衡
- **高性能通信**: 基于 gRPC 的高效节点间通信
- **可观测性**: 内置追踪和监控支持
- **容错机制**: 内置故障恢复和错误处理机制
- **发布订阅**: 支持基于主题的消息发布和订阅

## 适用场景

- **游戏服务器**: 适用于需要处理大量并发用户和实时通信的游戏服务
- **物联网应用**: 处理大规模设备连接和消息路由
- **微服务架构**: 构建可扩展的分布式服务系统
- **实时数据处理**: 处理高并发的数据流和事件流
- **分布式计算**: 支持复杂的分布式计算任务

## 优势

1. **简单易用**: 提供直观的 API，降低分布式系统开发难度
2. **高性能**: 基于 Go 的高并发特性，提供卓越的性能表现
3. **可扩展**: 支持水平扩展，轻松应对业务增长
4. **可靠性**: 内置故障恢复机制，提高系统稳定性
5. **开发效率**: 提供完整的工具集，加速开发周期

</br>

### 1. 快速开始
> 使用 braid-cli 工具安装脚手架项目

> 一个最小可工作游戏服务器，作为您使用 Braid 的起点

```shell
# 1. Install CLI Tool
$ go install github.com/pojol/braid-cli@latest

# 2. Using the CLI to Generate a New Empty Project
$ braid-cli new "you-project-name" v0.1.7

# 3. Creating .go Files from Actor Template Configurations
$ cd you-project-name/template
$ go generate

# 4. Navigate to the services directory, then try to build and run the demo
$ cd you-project-name/node
$ go run main.go
```


### 2. 为 actor 添加一个事件具柄

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

### 3. 消息发送

```golang
m := msg.NewBuilder(context.TODO())
m.WithReqCustomFields(fields.RoomID(b.RoomID))
ctx.Call(b.ID, template.ACTOR_USER, constant.Ev_UpdateUserInfo, m.Build())
```


### 4. 默认支持 jaeger 链路追踪
[![image.png](https://i.postimg.cc/wTVhQhyM/image.png)](https://postimg.cc/XprGVBg6)


<div style="display: flex; align-items: center; margin: 1em 0;">
  <div style="flex-grow: 1; height: 1px; background-color: #ccc;"></div>
  <div style="margin: 0 10px; font-weight: bold; color: #666;">测试机器人</div>
  <div style="flex-grow: 1; height: 1px; background-color: #ccc;"></div>
</div>

### 5. 通过测试机器人验证 braid 提供的服务器接口
> 使用上面的脚手架工程

```shell
$ cd you-project-name/testbots

# 1. 运行机器人服务器
$ go run main.go

# 2. 下载 gobot 编辑器（最新版本
https://github.com/pojol/gobot/releases

# 3. 运行 gobot 编辑器
$ run gobot_editor_[ver].exe or .dmg

# 4. 进入到 bots 页签
# 5. 将 testbots 目录中的 testbot.bh 文件拖拽到 bots 页面中
# 6. 选中 testbot 机器人，点击 load 加载 testbot
# 7. 点击左下角按钮，构建机器人实例
# 8. 点击单步运行按钮，查看机器人和 braid 服务器交互情形
```

[测试机器人 Gobot](https://github.com/pojol/gobot)
[![image.png](https://i.postimg.cc/LX5gbV34/image.png)](https://postimg.cc/xJrdkMZB)

---

### benchmark
```shell
goos: darwin
goarch: amd64
cpu: VirtualApple @ 2.50GHz
```
| 测试项 | 节点数量 | 性能指标 |
|--------|----------|----------|
| [dynamic-picker](https://github.com/pojol/braid/blob/master/tests/addressbook_test.go) | 10 | 500 actors/s |
| [call](https://github.com/pojol/braid/blob/master/tests/call_benchmark_test.go) | 2 (a1 -> a2 -> b1) | 14000 calls/s |