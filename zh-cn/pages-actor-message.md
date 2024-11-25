# 消息处理

</br>

* [消息处理](#注册消息处理chain)
* [Timerhandler](#注册-timerhandler)
* [MQ](#订阅消息并注册处理函数)
* [中间件](#使用中间件解析客户端消息)

</br>

### 注册消息处理Chain
> 下面的代码是一个标准的消息处理函数具柄，它会被注册到 actor 的消息映射表中；actor 的 runtime 会在接收到新的消息后顺序的去调用对应的处理具柄，在 actor 中任何的消息处理函数都是`同步`的！
```go
type UserStateType struct{}

func MakeUserUseItem(ctx core.ActorContext) core.IChain {

	return &actor.DefaultChain{
		Handler: func(mw *msg.Wrapper) error {

			state := ctx.GetValue(UserStateType{}).(*user.EntityWrapper)

			return nil
		},
	}

}
```
* **ctx** - actor提供的系统能力被集成到了 ctx 中，便于用户进行使用
* **state** - 通过 ctx 可以提取出 state 指针，用于操作 actor 的 state 对象（注：通常在 init 中进行注入
* **ichain** - chain 抽象了一组消息的处理函数
* **mw** - 消息体，在一个消息的处理链路中，这个消息体是会被一直传递下去的（注：包括跨服务的处理链路

</br>

### 注册 Timerhandler
> 为 actor 注册一个 timerhandler， 这个处理同样都是同步逻辑；下面的示例是创建了一个定时同步的 timer 用于将 state 同步到 cache
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

* **dueTime** timer 的启动延迟时间，这里填0表示立即启动
* **interval** 每次 tick 的间隔时间（毫秒），这里是1分钟一次
* **callback** timer 的回调函数
* **args** timer 回调函数的参数

</br>

### 订阅消息并注册处理函数
> 有时候我们在消息传递时，需要一些异步或离线机制，比如聊天频道中的离线消息，我们可以先将消息缓存在队列（mq)中，等待 actor 实例化之后再进行处理，这种情形就可以使用订阅；
```go
a.SubscriptionEvent(events.EvChatMessageStore, a.Id, func() {
  a.RegisterEvent(events.EvChatMessageStore, events.MakeChatStoreMessage)
}, pubsub.WithTTL(time.Hour*24*30))
```

* **SubscriptionEvent** 订阅 EvChatMessageStore 这个主题，并创建一个id为 a.Id 的 channel
* **WithTTL** 设置这个主题的消息过期时间
* **WithLimit** 设置这个队列的最大消息数量
* **Callback** 当从队列中获取到新的消息时，这个消息会被路由到 callback 中的实际处理函数具柄

<div style="display: flex; align-items: center; margin: 1em 0;">
  <div style="flex-grow: 1; height: 1px; background-color: #ccc;"></div>
  <div style="margin: 0 10px; font-weight: bold; color: #666;">进阶</div>
  <div style="flex-grow: 1; height: 1px; background-color: #ccc;"></div>
</div>

### 使用中间件解析客户端消息
> 在 chain 中，我们可以在 before 或 after 置入消息中间件处理一些通用逻辑；类似下面这个消息处理函数，我们添加了一个 msgunpack 的中间件，用于通用的消息解包逻辑（顺便对入参进行了打印

```go
func HttpHello(ctx core.ActorContext) core.IChain {

	unpackCfg := &middleware.MessageUnpackCfg[*gameproto.HelloReq]{}

	return &actor.DefaultChain{
		Before: []actor.EventHandler{middleware.MessageUnpack(unpackCfg)},
		Handler: func(mw *msg.Wrapper) error {

			req := unpackCfg.Msg.(*gameproto.HelloReq)

			return nil
		},
	}

}
```

* 中间件
> 抽象一些可复用的消息处理函数

```go
type MessageUnpackCfg[T any] struct {
	MsgTy T
	Msg   interface{}
}

func MessageUnpack[T any](cfg *MessageUnpackCfg[T]) actor.EventHandler {
	return func(msg *msg.Wrapper) error {
		var msgInstance proto.Message
		msgType := reflect.TypeOf(cfg.MsgTy)

		// 检查是否为指针类型
		if msgType.Kind() == reflect.Ptr {
			msgInstance = reflect.New(msgType.Elem()).Interface().(proto.Message)
		} else {
			msgInstance = reflect.New(msgType).Interface().(proto.Message)
		}

		// 解析消息
		err := proto.Unmarshal(msg.Req.Body, msgInstance)
		if err != nil {
			return fmt.Errorf("unpack msg err %v", err.Error())
		}

		// 打印消息类型和字段信息
		log.InfoF("[req event] actor_id : %s actor_ty : %s event : %s: params : %s",
			msg.Req.Header.TargetActorID,
			msg.Req.Header.TargetActorType,
			reflect.TypeOf(msgInstance).Elem().Name(), printMessageFields(msgInstance))

		// 将解析后的消息赋值给 cfg.Msg
		cfg.Msg = msgInstance

		return nil
	}
}
```