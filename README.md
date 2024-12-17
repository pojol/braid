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
> Braid adopts a minimalist design philosophy - with just three core concepts and six basic interfaces, you can build any single-node or distributed game server architecture

|  |  | | |
|-| ------ | ------ | ------- |
| Core Concepts | **Actor** | **Handler** | **State** |
|| ------ | ------ | ------- |
| Message Sending |[Call](https://pojol.github.io/braid/#/pages-actor-send)| `Send`| `Pub`|
| Event Subscription |[OnEvent](https://pojol.github.io/braid/#/pages-actor-message)| `OnTimer`| `Sub`|

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
