# 实现 chat 的事件逻辑

</br>

* [EvChatRouterMessage](#消息路由)
* [EvChatChannelAddUser](#添加用户)
* [EvChatChannelRmvUser](#移除用户)
* [EvChatChannelReceived](#接收到新的消息)
* [EvChatMessageStore](#离线消息)

</br>

### 消息体
> 消息的基础结构
```protobuf
message ChatMessage {	
    string SenderID = 1;    // 消息发送者
    string ReceiverID = 2;	// 消息接收者
    string Content = 3;		// 消息内容
    int64 Time = 4;			// 发送时间（unix时间戳
    string Channel = 5;		// 消息所在的频道
    map<string, string> Meta = 6; // 自定义描述结构（工会名，地域，称号
}
```

</br>

### 消息路由
> EvChatRouterMessage - 接受客户端所有的聊天消息，通过 header 路由到不同的 chat channel 中
```go
func MakeChatSendCmd(ctx core.ActorContext) core.IChain {

	unpackCfg := &middleware.MessageUnpackCfg[*gameproto.ChatSendReq]{}

	return &actor.DefaultChain{
		Before: []actor.EventHandler{middleware.MessageUnpack(unpackCfg)},
		Handler: func(mw *msg.Wrapper) error {

			req := unpackCfg.Msg.(*gameproto.ChatSendReq)

            if req.Msg.Channel == constant.ChatPrivateChannel { // 私聊需要特殊处理
				if ctx.AddressBook().Exist(req.Msg.ReceiverID) {
					// 检查用户在线，则直接发送到这个用户的私聊频道
					ctx.Call(router.Target{ID: req.Msg.ReceiverID, Ev: EvChatChannelReceived}, mw)
				} else {
					// 离线将 聊天信息 存储在目标用户的离线channel中，等待用户上线后消费
					ctx.Pub(EvChatMessageStore, "private_"+req.Msg.ReceiverID, msg)
				}
			} else {
				// 直接路由到频道，但可以校验下频道名是否合法
				ctx.Call(router.Target{ID: def.SymbolLocalFirst, Ty: req.Msg.Channel, Ev: EvChatChannelReceived}, mw)
			}

			return nil
		},
	}
}
```

</br>

### 接收到新的消息
> EvChatChannelReceived - 频道接受到新的消息，并广播给频道内的其他玩家
```go
func MakeChatRecved(ctx core.ActorContext) core.IChain {

	unpackCfg := &middleware.MessageUnpackCfg[*gameproto.ChatSendReq]{}

	return &actor.DefaultChain{
		Before: []actor.EventHandler{middleware.MessageUnpack(unpackCfg)},
		Handler: func(mw *msg.Wrapper) error {

			req := unpackCfg.Msg.(*gameproto.ChatSendReq)
			state := ctx.GetValue(ChatStateType{}).(*chat.State)

            // 记录到 state 中
			state.AppendMsg(req.Msg)

			notify := gameproto.ChatMessageNotify{
				MsgLst: []*gameproto.ChatMessage{
					req.Msg,
				},
			}

			mw.Res.Body, _ = proto.Marshal(&notify)

			for _, v := range state.Users {

				if v.ActorID == req.Msg.SenderID {
					continue
				}

				mw.Res.Header.Token = v.ActorToken
				mw.Res.Header.Event = EvChatMessageNty

				ctx.Send(router.Target{
					ID: v.ActorGate,
					Ty: config.ACTOR_WEBSOCKET_ACCEPTOR,
					Ev: EvWebsoketNotify,
				}, mw)
			}

			return nil
		},
	}
}
```

### 添加用户 & 删除用户
> 从 state 中添加或删除某个用户
```go
// 略 - 直接调用 state 的接口即可
```

</br>

### 离线消息
> EvChatMessageStore - 处理私聊用户不在线的情况（消费离线聊天消息队列中的消息
```go
// 用户 actor 订阅私聊频道的离线消息
func (a *userActor) Init(ctx context.Context) {
	// ...

	// 订阅自己的私聊 "private_"+a.Id
	// pubsub.WithTTL 这个频道内的消息 一个月后过期删除
	a.Sub(events.EvChatMessageStore, "private_"+a.Id, events.MakeChatStoreMessage, pubsub.WithTTL(time.Hour*24*30))
}

// handler
func MakeChatStoreMessage(ctx core.ActorContext) core.IChain {
	return &actor.DefaultChain{
		Handler: func(mw *msg.Wrapper) error {

            state := ctx.GetValue(ChatStateType{}).(*chat.State)

			state.AppendMsg(mw.req.msg)

            // 发送给自己
            ctx.Send(router.Target{ID: ctx.ID(), Ev: EvWebsoketNotify}, mw)

			return nil
		},
	}
}
```
