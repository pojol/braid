# Braid 一个轻量级的 Actor 游戏开发框架
> Braid 是一个以 Actor 模型为核心驱动的创新型无服务器游戏框架。它通过统一的寻址系统实现自动负载管理，使开发者能够专注于设计和实现，而无需关心复杂的分布式系统组件。

[![Go Report Card](https://goreportcard.com/badge/github.com/pojol/braid?style=flat-square)](https://goreportcard.com/report/github.com/pojol/braid)
[![Documentation](https://img.shields.io/badge/Documentation-Available-brightgreen)](https://pojol.github.io/braid/#/)
<!--
[![Discord](https://img.shields.io/discord/1210543471593791488?color=7289da&label=Discord&logo=discord&style=flat-square)](https://discord.gg/yXJgTrkWxT)
-->

[![image.png](https://i.postimg.cc/BbvzLhfN/image.png)](https://postimg.cc/Vr3g2W6b)

### Features
> braid 采用简洁的设计理念，只需通过下面的三个核心概念和六个基础接口，即可构建各类单点或分布式游戏服务器架构

|  |  | | |
|-| ------ | ------ | ------- |
| 核心概念 | `Actor` | `Handler` | `State` |
|| ------ | ------ | ------- |
| 消息发送 |`Call`| `Send`| `Pub`|
| 事件订阅 |`OnEvent`| `OnTimer`| `Sub`|

- `Actor` 表示在集群中的计算单元，负责维护接收消息的handlers和状态，通常 actor = 一系列计算函数和状态的集合，比如 user, mail, rank, chat ...
- `Handler` 表示处理具体消息的函数，可以是事件处理，也可以是定时器处理，也可以是消息订阅（在handle中的逻辑都可以认为是同步的，不需要担心异步逻辑）
- `State` 是 actor 的状态，用于存储和读写 actor 的数据
- `Call` 表示一次阻塞调用（调用可以传入路由规则，直接发送，广播发送，随机发送 等...
- `Send` 表示一次非阻塞调用
- `Pub` 表示向 MQ 的某个 Topic 发送一条消息
- `OnEvent` 注册一个新的事件处理函数
- `OnTimer` 注册一个新的 timer
- `Sub` 表示订阅某个 Topic 的 Channel

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
