# process message

* [Message Processing](#register-message-handler)
* [Timerhandler](#register-timerhandler)
* [MQ](#subscribe-message-and-register-handler)
* [Middleware](#use-middleware-to-parse-client-message)

</br>

### register message handler
> The following code is a standard message processing function handle that will be registered in the actor's message mapping table. The actor's runtime will sequentially call the corresponding handling handles after receiving new messages. In an actor, all message processing functions are `synchronous`!
```go
type UserStateType struct{}

func MakeUserUseItem(ctx core.ActorContext) core.IChain {

	return &actor.DefaultChain{
		Handler: func(mw *router.MsgWrapper) error {

			state := ctx.GetValue(UserStateType{}).(*user.EntityWrapper)

			return nil
		},
	}

}
```
* **ctx** - The system capabilities provided by the actor are integrated into ctx for user convenience
* **state** - Through ctx, the state pointer can be extracted to operate on the actor's state object (Note: usually injected in init)
* **ichain** - chain abstracts a group of message processing functions
* **mw** - Message body, which will be passed along the message processing chain (Note: including cross-service processing chains)

</br>

### register Timerhandler
> Register a timerhandler for the actor, which is also synchronous logic. The following example creates a timer for periodic synchronization to sync the state to the cache
```go
dueTime := 0
interval := 1000*60
callback := func(interface{}) error {
  a.entity.Sync(context.TODO())
  return nil
}
args := interface{}

a.RegisterTimer(dueTime, interval, callback, args)
```

* **dueTime** The startup delay time of the timer, 0 here means start immediately
* **interval** The interval time for each tick (milliseconds), here it's once per minute
* **callback** The callback function of the timer
* **args** Parameters for the timer callback function

</br>

### subscribe message and register handler
> Sometimes when passing messages, we need some asynchronous or offline mechanisms, such as offline messages in chat channels. We can first cache the messages in a queue (mq) and process them after the actor is instantiated. This scenario can use subscriptions:
```go
a.SubscriptionEvent(events.EvChatMessageStore, a.Id, func() {
  a.RegisterEvent(events.EvChatMessageStore, events.MakeChatStoreMessage)
}, pubsub.WithTTL(time.Hour*24*30))
```

* **SubscriptionEvent** Subscribe to the EvChatMessageStore topic and create a channel with id a.Id
* **WithTTL** Set the message expiration time for this topic
* **WithLimit** Set the maximum number of messages for this queue
* **Callback** When a new message is retrieved from the queue, it will be routed to the actual handling function handle in the callback

<div style="display: flex; align-items: center; margin: 1em 0;">
  <div style="flex-grow: 1; height: 1px; background-color: #ccc;"></div>
  <div style="margin: 0 10px; font-weight: bold; color: #666;">advanced</div>
  <div style="flex-grow: 1; height: 1px; background-color: #ccc;"></div>
</div>

### use middleware to parse client message
> In the chain, we can insert message middleware in before or after to handle some common logic. Similar to the following message processing function, we added a msgunpack middleware for general message unpacking logic (and printed the input parameters)

```go
func HttpHello(ctx core.ActorContext) core.IChain {

	unpackCfg := &middleware.MessageUnpackCfg[*gameproto.HelloReq]{}

	return &actor.DefaultChain{
		Before: []actor.EventHandler{middleware.MessageUnpack(unpackCfg)},
		Handler: func(mw *router.MsgWrapper) error {

			req := unpackCfg.Msg.(*gameproto.HelloReq)

			return nil
		},
	}

}
```

* Middleware
> Abstract some reusable message processing functions

```go
type MessageUnpackCfg[T any] struct {
	MsgTy T
	Msg   interface{}
}

func MessageUnpack[T any](cfg *MessageUnpackCfg[T]) actor.EventHandler {
	return func(msg *router.MsgWrapper) error {
		var msgInstance proto.Message
		msgType := reflect.TypeOf(cfg.MsgTy)

		// Check if it's a pointer type
		if msgType.Kind() == reflect.Ptr {
			msgInstance = reflect.New(msgType.Elem()).Interface().(proto.Message)
		} else {
			msgInstance = reflect.New(msgType).Interface().(proto.Message)
		}

		// Parse the message
		err := proto.Unmarshal(msg.Req.Body, msgInstance)
		if err != nil {
			return fmt.Errorf("unpack msg err %v", err.Error())
		}

		// Print message type and field information
		log.InfoF("[req event] actor_id : %s actor_ty : %s event : %s: params : %s",
			msg.Req.Header.TargetActorID,
			msg.Req.Header.TargetActorType,
			reflect.TypeOf(msgInstance).Elem().Name(), printMessageFields(msgInstance))

		// Assign the parsed message to cfg.Msg
		cfg.Msg = msgInstance

		return nil
	}
}
```