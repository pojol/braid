# 设计一个聊天服务器

</br>

* [设计](#设计)
* [聊天的数据模型](#聊天的数据模型)
* [ChatActor](#聊天-actor)
* [构建ChatActor](#构建-chatactor)
* [功能逻辑:消息是如何被路由到各个ChatActor的](#消息是如何被路由到各个chatactor的)
* [功能逻辑:聊天消息处理](#聊天消息处理)
* [功能逻辑:离线聊天消息处理](#离线聊天消息处理)

### 设计
> 提前概括一些逻辑，用于抽象设计
1. 为玩家actor添加 进入/离开频道 的API
    * 我们需要玩家有自由进入和离开频道的能力，在游戏中聊天服务本质上就是提供频道服务
2. 设计一个 频道actor 用于处理聊天逻辑
    * 每个频道都是一个单独的 actor 但他们的权重可能不一样
    * 每个平台需要有状态信息，用于管理频道中的数据（用户，和内容
    * 频道需要有广播能力，通知给订阅这个频道的玩家
3. 离线消息处理
    * 私聊频道可能还需要有离线存储能力（因为很多时候并不能保证目标玩家在线
4. 设计一个 消息路由 actor
    * 为了方便聊天消息处理，用一个统一的聊天消息路由器去处理消息的重定向

</br>

### 聊天的数据模型
```go
type State struct {
    // 频道名称
	Channel string
    // 玩家的 session 信息
	Users   []comm.UserSession
	// 这个频道内的消息列表
	MsgHistory []gameproto.ChatMessage
}
```

### 聊天 Actor
```go
type chatChannelActor struct {
	*actor.Runtime
	state *chat.State
}

func NewChatActor(p core.IActorBuilder) core.IActor {
	return &chatChannelActor{
		Runtime: &actor.Runtime{Id: p.GetID(), Ty: p.GetOpt("channel").(string), Sys: p.GetSystem()},
		state: &chat.State{
			Channel: p.GetOpt("channel").(string),
		},
	}
}

func (a *chatChannelActor) Init(ctx context.Context) {
	a.Runtime.Init(ctx)

	a.Context().WithValue(events.ChatStateType{}, a.state)

	a.RegisterEvent(events.EvChatChannelReceived, events.MakeChatRecved)
	a.RegisterEvent(events.EvChatChannelAddUser, events.MakeChatAddUser)
	a.RegisterEvent(events.EvChatChannelRmvUser, events.MakeChatRemoveUser)

	err := a.SubscriptionEvent(events.EvChatMessageStore, a.Id, func() {
		a.RegisterEvent(events.EvChatMessageStore, events.MakeChatStoreMessage)
	}, pubsub.WithTTL(time.Hour*24*30))
	if err != nil {
		log.WarnF("actor %v ty %v subscription event %v err %v", a.Id, a.Ty, events.EvChatMessageStore, err.Error())
	}
}
```

* 主要细节
1. 置入 state 到 actor
2. 构建的时候 将 channel 名称填入（这个 actor 可以用于各种类型的 chat actor
3. 在 init 内
    * 绑定 state 到 ctx
    * 绑定 add / remove user 的事件处理
    * 绑定 接收到消息 的事件处理
    * 绑定离线消息的处理（并设置消息的超时时间

</br>

### 构建 ChatActor
> 对于静态的 chat actor 比如 工会聊天频道，全服聊天频道，地区聊天频道等，通过配置进行构建
```yaml
  actors:
    - name: "CHAT"
      options:
        channel: "global"   # 全服聊天频道
        weight: 10000       # 大一些，如果用户数多可以独占一个节点
    - name: "CHAT"
      options:
        channel: "guild"    # 工会聊天频道
        weight: 100
```

> 动态构建的聊天频道(私聊，一个用户附带一个（在 user actor 构建成功后创建
```go
userActor.Sys.Loader(actor_types.CHAT).
    WithID("chat."+constant.ChatPrivateChannel+"."+a.Id).
    WithOpt("channel", constant.ChatPrivateChannel).
    WithOpt("actorID", a.Id).WithPicker().Build()
}
```

</br>

### 消息是如何被路由到各个ChatActor的
> 设计一个 ChatRouterActor, 无状态，只负责转发逻辑（所以只需要绑定下面的实现函数即可， 主要的逻辑就是转发消息
```go
func MakeChatSendCmd(ctx core.ActorContext) core.IChain {

	unpackCfg := &middleware.MessageUnpackCfg[*gameproto.ChatSendReq]{}

	return &actor.DefaultChain{
		Before: []actor.EventHandler{middleware.MessageUnpack(unpackCfg)},
		Handler: func(mw *msg.Wrapper) error {

			req := unpackCfg.Msg.(*gameproto.ChatSendReq)

            if req.Msg.Channel == constant.ChatPrivateChannel { // 私聊需要特殊处理
				if ctx.AddressBook().Exist(req.Msg.ReceiverID) {
					ctx.Call(router.Target{ID: "ReceiverChatID", Ev: EvChatChannelReceived}, mw)
				} else { // 离线将 聊天信息 存储在目标用户的离线channel中，等待用户上线后消费
					ctx.Pub(EvChatMessageStore, msg)
				}
			} else {
				ctx.Call(router.Target{ID: def.SymbolLocalFirst, Ty: "channel_type", Ev: EvChatChannelReceived}, mw)
			}

			return nil
		},
	}
}
```


### 聊天消息处理
> 消息处理的主要逻辑是，将消息存储在频道的消息队列中，并将刚刚接收到的消息通知给频道内的玩家
```go
func MakeChatRecved(ctx core.ActorContext) core.IChain {

	unpackCfg := &middleware.MessageUnpackCfg[*gameproto.ChatSendReq]{}

	return &actor.DefaultChain{
		Before: []actor.EventHandler{middleware.MessageUnpack(unpackCfg)},
		Handler: func(mw *msg.Wrapper) error {

			req := unpackCfg.Msg.(*gameproto.ChatSendReq)
			state := ctx.GetValue(ChatStateType{}).(*chat.State)

            // 存起来
			state.MsgHistory = append(state.MsgHistory, *req.Msg)

			notify := gameproto.ChatMessageNotify{
				MsgLst: []*gameproto.ChatMessage{
					req.Msg,
				},
			}

			mw.Res.Body, _ = proto.Marshal(&notify)

			if req.Msg.Channel == constant.ChatPrivateChannel {
				ctx.Send(router.Target{ID: def.SymbolLocalFirst, Ty: config.ACTOR_WEBSOCKET_ACCEPTOR, Ev: EvWebsoketNotify},
					mw,
				)
			} else {

				for _, v := range state.Users {

					mw.Res.Header.Token = v.ActorToken
					mw.Res.Header.Event = EvChatMessageNty

					ctx.Send(router.Target{
						ID: v.ActorGate,
						Ty: config.ACTOR_WEBSOCKET_ACCEPTOR,
						Ev: EvWebsoketNotify,
					},
						mw,
					)
				}

			}

			return nil
		},
	}
}
```

</br>

### 离线聊天消息处理
> 离线聊天处理主要分两部分 1. 在构建一个 user actor 时，同时需要构建一个它的专属 private chat actor， 2. 在玩家自己的 private chat actor 启动时，从 topic 中拿出离线聊天数据，并同步给玩家

1. 动态构建 private chat actor

```go
func (a *UserActor) Init(ctx context.Context) {

// 用户构建完成时，构建一个专属的 private chat actor
a.Sys.Loader(config.ACTOR_PRIVATE_CHAT).
    WithID("chat."+constant.ChatPrivateChannel+"."+a.Id).
    WithOpt("channel", constant.ChatPrivateChannel).
    WithOpt("actorID", a.Id).WithPicker().Build()

}
```

2. 离线消息处理

```go
func MakeChatStoreMessage(ctx core.ActorContext) core.IChain {
	return &actor.DefaultChain{
		Handler: func(mw *msg.Wrapper) error {

            offlineMsg := mw.req.msg

            state := ctx.GetValue(ChatStateType{}).(*chat.State)

            // 存起来
			state.MsgHistory = append(state.MsgHistory, *req.Msg)

            // 广播给用户自己
            ctx.Send(router.Target{ID: "actor_id", Ev: EvWebsoketNotify}, mw)

			return nil
		},
	}
}
```

---

### [完整的实现](https://github.com/pojol/braid-demo/tree/master/demos/3_chat)