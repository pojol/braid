# 发送消息

* [消息发送机制](#消息发送机制)
* [阻塞调用](#阻塞调用)
* [异步调用](#异步调用)
* [发布消息](#发布消息发送到mq)
* [可重入消息调用](#使用可重入消息调用)

</br>

### 消息发送机制
> 在 braid 中，用户不需要知道 actor 位于集群中的具体位置，可以通过4种发布模式，和不同的查找模式，进行消息发送；
* 发布模式
  * 阻塞调用(Call - 这个调用会让当前处理函数阻塞，直到调用完成，或超时失败
  * 异步调用(Send - 发送后即往下执行，没有返回值（通常可以用于一些耗时较长的计算逻辑
  * 发布消息(Pub - 将消息发布到 mq，等待订阅者消费
  * 可重入调用(ReenterCall - 可以为调用设置钩子，在异步执行完后执行其他的同步动作（可以链式调用
* 消息查找模式
  * 按id - 直接通过 id 进行查找
  * 按符号 --
  * SymbolWildcard ("?" - 随机分配到任意一个指定同类型 actor 上
  * SymbolGroup ("#" - 分配到 group 列表中的 actor 上
  * SymbolAll ("*" - 分配到集群中所有同类型 actor 上
  * SymbolLocalFirst ("~" - 随机分配到任意一个指定同类型 actor 上（但优先本节点

</br>

### 阻塞调用
> 通过 ctx.Call 发起一次阻塞调用，目标 actor 可以位于集群的任意位置；

```go
func MakeChatRecved(ctx core.ActorContext) core.IChain {

	unpackCfg := &middleware.MessageUnpackCfg[*gameproto.ChatSendReq]{}

  return &actor.DefaultChain{
      Before: []actor.EventHandler{middleware.MessageUnpack(unpackCfg)},
      Handler: func(mw *msg.Wrapper) error {

      req := unpackCfg.Msg.(*gameproto.ChatSendReq)
      state := ctx.GetValue(ChatStateType{}).(*chat.State)

      ctx.Call(router.Target{
        ID: v.ActorGate,
        Ty: config.ACTOR_WEBSOCKET_ACCEPTOR,
      }, mw)

      return nil
    },
  }
}

```

</br>

### 异步调用
> 和阻塞调用一样，只需替换接口即可
```go
ctx.Send(router.Target{
  ID: v.ActorGate,
  Ty: config.ACTOR_WEBSOCKET_ACCEPTOR,
}, mw)
```

</br>

### 发布消息（发送到MQ
> 将消息发布到 topic， 如果这个 topic 需要 1:1 消费模型则只需创建一个消费者， 如果是 1:M 则创建多个消费者
```go
ctx.Pub(topic, msg)
```

<div style="display: flex; align-items: center; margin: 1em 0;">
  <div style="flex-grow: 1; height: 1px; background-color: #ccc;"></div>
  <div style="margin: 0 10px; font-weight: bold; color: #666;">进阶</div>
  <div style="flex-grow: 1; height: 1px; background-color: #ccc;"></div>
</div>

### 使用可重入消息调用
> 可重入消息，可用于一些需要访问3方 api 后进行处理的业务逻辑； 如刷新 token 后将新的 token 设置到 state
```go
// 发起一次异步可重入调用
future := ctx.ReenterCall(mw.Ctx, router.Target{}, mw)

// 注册回调函数，在异步调用完成后处理结果， 注意：
//  1. Then 方法本身是同步调用，立即返回。
//  2. future 代表的异步操作是并行执行的。
//  3. 回调函数会在 future 完成后被依次同步调用
future.Then(func(i interface{}, err error) {

}).Then(func(i interface{}, err error) {
// 链式调用
})

// 注意：这里立即返回，不等待异步操作完成
return nil
```


