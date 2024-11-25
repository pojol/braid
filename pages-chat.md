# Designing a Chat Server

</br>

* [Design](#design)
* [Chat Data Model](#chat-data-model)
* [ChatActor](#chat-actor)
* [Building ChatActor](#building-chatactor)
* [Functionality: How Messages are Routed to ChatActors](#how-messages-are-routed-to-chatactors)
* [Functionality: Chat Message Processing](#chat-message-processing)
* [Functionality: Offline Chat Message Processing](#offline-chat-message-processing)

### Design
> Summarizing some logic in advance for abstract design
1. Entering/Leaving Channels
    * We need players to have the ability to freely enter and leave channels. In games, chat service is essentially providing channel services.
2. Channel Logic
    * Each channel is a separate actor, but they may have different weights.
    * Each platform needs to have state information to manage data in the channel (users and content).
    * Channels need to have broadcasting capabilities to notify players subscribed to this channel.
3. Offline Message Processing
    * Private chat channels may also need offline storage capabilities (because often the target player cannot be guaranteed to be online).
4. Message Routing
    * To facilitate chat message processing, use a unified chat message router to handle message redirection.

</br>

### Chat Data Model
```go
type State struct {
    // Channel name
	Channel string
    // Player session information
	Users   []comm.UserSession
	// Message list in this channel
	MsgHistory []gameproto.ChatMessage
}
```

### Chat Actor
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

* Key details
1. Insert state into actor
2. Fill in the channel name when building (this actor can be used for various types of chat actors)
3. In init
    * Bind state to ctx
    * Bind event handlers for add/remove user
    * Bind event handler for received messages
    * Bind offline message processing (and set message timeout)

</br>

### Building ChatActor
> For static chat actors such as guild chat channels, global chat channels, regional chat channels, etc., build through configuration
```yaml
  actors:
    - name: "CHAT"
      options:
        channel: "global"   # Global chat channel
        weight: 10000       # Larger, can occupy a separate node if there are many users
    - name: "CHAT"
      options:
        channel: "guild"    # Guild chat channel
        weight: 100
```

> Dynamically built chat channels (private chat, one per user, created after the user actor is successfully built)
```go
userActor.Sys.Loader(actor_types.CHAT).
    WithID("chat."+constant.ChatPrivateChannel+"."+a.Id).
    WithOpt("channel", constant.ChatPrivateChannel).
    WithOpt("actorID", a.Id).WithPicker().Build()
}
```

</br>

### How Messages are Routed to ChatActors
> Design a ChatRouterActor, stateless, only responsible for forwarding logic (so only need to bind the implementation function below). The main logic is just forwarding messages
```go
func MakeChatSendCmd(ctx core.ActorContext) core.IChain {

	unpackCfg := &middleware.MessageUnpackCfg[*gameproto.ChatSendReq]{}

	return &actor.DefaultChain{
		Before: []actor.EventHandler{middleware.MessageUnpack(unpackCfg)},
		Handler: func(mw *msg.Wrapper) error {

			req := unpackCfg.Msg.(*gameproto.ChatSendReq)

            if req.Msg.Channel == constant.ChatPrivateChannel { // Special handling for private chat
				if ctx.AddressBook().Exist(req.Msg.ReceiverID) {
					ctx.Call(router.Target{ID: "ReceiverChatID", Ev: EvChatChannelReceived}, mw)
				} else { // Store chat information in the target user's offline channel for consumption when the user comes online
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

### Chat Message Processing
> The main logic of message processing is to store messages in the channel's message queue and notify players in the channel of the newly received messages
```go
func MakeChatRecved(ctx core.ActorContext) core.IChain {

	unpackCfg := &middleware.MessageUnpackCfg[*gameproto.ChatSendReq]{}

	return &actor.DefaultChain{
		Before: []actor.EventHandler{middleware.MessageUnpack(unpackCfg)},
		Handler: func(mw *msg.Wrapper) error {

			req := unpackCfg.Msg.(*gameproto.ChatSendReq)
			state := ctx.GetValue(ChatStateType{}).(*chat.State)

            // Store it
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

### Offline Chat Message Processing
> Offline chat processing is mainly divided into two parts: 1. When building a user actor, a dedicated private chat actor needs to be built for it. 2. When the player's own private chat actor starts, it takes offline chat data from the topic and synchronizes it to the player.

1. Dynamically build private chat actor

```go
func (a *UserActor) Init(ctx context.Context) {

// When the user is built, build a dedicated private chat actor
a.Sys.Loader(config.ACTOR_PRIVATE_CHAT).
    WithID("chat."+constant.ChatPrivateChannel+"."+a.Id).
    WithOpt("channel", constant.ChatPrivateChannel).
    WithOpt("actorID", a.Id).WithPicker().Build()

}
```

2. Offline message processing

```go
func MakeChatStoreMessage(ctx core.ActorContext) core.IChain {
	return &actor.DefaultChain{
		Handler: func(mw *msg.Wrapper) error {

            offlineMsg := mw.req.msg

            state := ctx.GetValue(ChatStateType{}).(*chat.State)

            // Store it
			state.MsgHistory = append(state.MsgHistory, *req.Msg)

            // Broadcast to the user themselves
            ctx.Send(router.Target{ID: "actor_id", Ev: EvWebsoketNotify}, mw)

			return nil
		},
	}
}
```

---

### [Complete Implementation](https://github.com/pojol/braid-demo/tree/master/demos/3_chat)