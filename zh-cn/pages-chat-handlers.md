# 实现 chat 的事件逻辑

* [EvChatRouterMessage](#消息路由)
* [EvChatChannelAddUser](#添加用户)
* [EvChatChannelRmvUser](#移除用户)
* [EvChatChannelReceived](#接收到新的消息)
* [EvChatMessageStore](#存储离线消息)

### 消息路由
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


### 接收到新的消息
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

### 存储离线消息
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
