# send message

* [Message Sending Mechanism](#message-sending-mechanism)
* [Blocking Call](#blocking-call)
* [Asynchronous Call](#asynchronous-call)
* [Publish Message](#publish-message-send-to-mq)
* [Reentrant Message Call](#using-reentrant-message-call)

</br>

### Message Sending Mechanism
> In braid, users don't need to know the specific location of an actor in the cluster. Messages can be sent through 4 publishing modes and different lookup modes:
* Publishing Modes
  * Blocking Call (Call) - This call will block the current processing function until the call is completed or times out
  * Asynchronous Call (Send) - Continues execution immediately after sending, with no return value (usually used for time-consuming computational logic)
  * Publish Message (Pub) - Publishes the message to MQ, waiting for subscribers to consume
  * Reentrant Call (ReenterCall) - Can set hooks for the call, executing other synchronous actions after asynchronous execution (can be chained)
* Message Lookup Modes
  * By ID - Direct lookup through ID
  * By Symbol --
  * SymbolWildcard ("?") - Randomly assigned to any specified actor of the same type
  * SymbolGroup ("#") - Assigned to actors in the group list
  * SymbolAll ("*") - Assigned to all actors of the same type in the cluster
  * SymbolLocalFirst ("~") - Randomly assigned to any specified actor of the same type (but prioritizes the local node)

</br>

### Blocking Call
> Initiate a blocking call through ctx.Call, the target actor can be located anywhere in the cluster:

```go
func MakeChatRecved(ctx core.ActorContext) core.IChain {

	unpackCfg := &middleware.MessageUnpackCfg[*gameproto.ChatSendReq]{}

  return &actor.DefaultChain{
      Before: []actor.EventHandler{middleware.MessageUnpack(unpackCfg)},
      Handler: func(mw *router.MsgWrapper) error {

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

````


</br>

### Asynchronous Call
> Same as blocking call, just replace the interface
````go
ctx.Send(router.Target{
  ID: v.ActorGate,
  Ty: config.ACTOR_WEBSOCKET_ACCEPTOR,
}, mw)
````


</br>

### Publish Message (Send to MQ)
> Publish the message to a topic. If this topic needs a 1:1 consumption model, only one consumer needs to be created. If it's 1:M, create multiple consumers

```go
ctx.Pub(topic, msg)
```


<div style="display: flex; align-items: center; margin: 1em 0;">
  <div style="flex-grow: 1; height: 1px; background-color: #ccc;"></div>
  <div style="margin: 0 10px; font-weight: bold; color: #666;">Advanced</div>
  <div style="flex-grow: 1; height: 1px; background-color: #ccc;"></div>
</div>

### Using Reentrant Message Call
> Reentrant messages can be used for business logic that needs to access 3rd party APIs before processing; such as setting a new token to the state after refreshing the token
````go
// Initiate an asynchronous reentrant call
future := ctx.ReenterCall(mw.Ctx, router.Target{}, mw)

// Register callback functions to handle results after the asynchronous call completes. Note:
//  1. The Then method itself is a synchronous call, returning immediately.
//  2. The asynchronous operation represented by future is executed in parallel.
//  3. Callback functions will be called synchronously in sequence after the future completes
future.Then(func(i interface{}, err error) {

}).Then(func(i interface{}, err error) {
// Chain call
})

// Note: This returns immediately, not waiting for the asynchronous operation to complete
return nil
````
