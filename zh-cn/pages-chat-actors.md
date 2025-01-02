# 设计一个聊天服务器

</br>

* [1.设计聊天所需的 actor](#设计聊天所需的-actor)
* [2.声明ChatActor](#声明ChatActor)
* [3.聊天依赖的数据项](#聊天依赖的数据项)
* [4.构建ChatActor](#构建-chatactor)

## 设计聊天所需的 actor
> 在编写代码之前，我们先梳理一下一个完备的 chat server 应该所需哪些功能项
1. 我们应该需要一个 channel 的概念，让用户可以在不同的分组中进行聊天
	* 创建，销毁
	* 加入，离开
	* 需要一个 channel 的 state 维护这个 channel 中的用户和聊天信息
2. 在整个集群中我们应该并行处理若干 channel，同时这些 channel 有的可能人多，有的可能人少，所以我们需要赋予 channel 不同的权重
3. 每个 channel 都需要有广播能力，对于全服聊天，对于广播来说可能需要特殊处理
4. 离线消息处理（私聊频道可能还需要有离线存储能力，因为很多时候并不能保证目标玩家在线
5. 需要一个消息路由，来对聊天消息进行重定向，这不属于其他 actor 的职责

> 因此我们需要设计如下的 actor

| Actor | 状态 |  描述 |
|-------|------|------|
| chat_channel_actor || 用于包装 channel 的逻辑 |
| | chat_channel_state | 用于包装 channel 的 state |
| chat_router_actor || 用于包装消息路由的逻辑 |

</br>


## 声明ChatActor
> chat channel actor 会被区分为三类（ private 表示私聊频道，一个用户持有一个， global 表示全服聊天， 自定义频道 用户可以自行创建

```go
type chatChannelActor struct {
	*actor.Runtime
	state *chat.State	// 持有 chat state
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

	// 将 state 绑定到 actor 的 context 中
	a.Context().WithValue(events.ChatStateType{}, a.state)

	// 绑定事件 - 接收到新的消息
	a.OnEvent(events.EvChatChannelReceived, events.MakeChatRecved)
	// 绑定事件 - 添加用户
	a.OnEvent(events.EvChatChannelAddUser, events.MakeChatAddUser)
	// 绑定事件 - 移除用户
	a.OnEvent(events.EvChatChannelRmvUser, events.MakeChatRemoveUser)

	// 绑定事件 - 存储离线消息
	a.Sub(events.EvChatMessageStore, a.Id, events.MakeChatStoreMessage, pubsub.WithTTL(time.Hour*24*30))
}
```

</br>

## 聊天依赖的数据项
* 频道内玩家信息
	* 拉黑用户列表
	* 玩家的 session 信息
	* 玩家的基础数据
* 给自定义房间添加密码
* 频道名
* 消息队列

```go
type ChatUser struct {
	UserID    string
	NickName  string
	Avatar    string

	SessionID string // 玩家的 session id

	BlackList []string // 黑名单列表
}

type State struct {
    // 频道名称（唯一，也可以表示类型
	Channel string
	// 密码 (自定义频道可以设置私密还是公开
	Password string

    // 频道内的玩家列表
	Users   []ChatUser
	// 这个频道内的消息列表
	MsgHistory []gameproto.ChatMessage
}
```

## 构建 ChatActor
> 对于静态的 chat actor 比如 工会聊天频道，全服聊天频道，地区聊天频道等，通过配置进行构建
```yaml
  actors:
    - name: "CHAT"
      options:
        channel: "global"   # 全服聊天频道
        weight: 10000       # 大一些，如果用户数多可以独占一个节点
		parallel: true		# 开启并行能力（主从
	- name: "Private"
	  options:
		channel: "private"  # 私聊频道(一个用户附带一个，在 user actor 构建成功后创建)
		weight: 100
    - name: "CHAT"
      options:
        channel: "custom"    # 自定义频道
        weight: 1000
		memberLimit: 1000 # 自定义频道的最大人数
```

</br>