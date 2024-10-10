# 为actor引入状态

</br>

* [绑定状态](#绑定状态)
* [设置和获取状态](#设置和获取状态)
* [通过 state 进行操作](#通过-state-进行操作)
---
* [提供给外部访问和修改](#提供给外部访问和修改-伪代码)
* [同步修改多个 actor 的 state](#需要同步修改多个-actor-的-state)

</br>

### 绑定状态

```go
type UserActor struct {
	*actor.Runtime
	state    *user.EntityWrapper    // 1. 将 state 绑定到 actor
}

func NewUserActor(p core.IActorBuilder) core.IActor {
	return &UserActor{
		Runtime:   &actor.Runtime{Id: p.GetID(), Ty: p.GetType(), Sys: p.GetSystem()},
		state:    user.NewEntityWapper(p.GetID()),  // 2. 构建 state
	}
}

func (a *UserActor) Init(ctx context.Context) {
    // 3. 装载 state
    err := a.state.Load(context.TODO())
	if err != nil {
		panic(fmt.Errorf("load user actor err %v", err.Error()))
	}
}

```

</br>

### 设置和获取状态

```go
func (a *UserActor) Init(ctx context.Context) {

    // 2. 将 state 装填到 context 中，方便 events 中获取
	a.Context().WithValue(events.UserStateType{}, a.entity) 
}

// events_user_base.go
type UserStateType struct{} // 1. 定义 state 在 context 中的类型

func MkOperatorXXX(ctx core.ActorContext) core.IChain {
    return &actor.DefaultChain{
		Handler: func(mw *router.MsgWrapper) error {

            // 3. 提取 state 用于各种修改和计算
			state := ctx.GetValue(UserStateType{}).(*user.EntityWrapper)
        }
    }
}

```

</br>

### 通过 state 进行操作

```go
func MkOperatorXXX(ctx core.ActorContext) core.IChain {
    return &actor.DefaultChain{
		Handler: func(mw *router.MsgWrapper) error {

            // 提取 state 用于各种修改和计算
			state := ctx.GetValue(UserStateType{}).(*user.EntityWrapper)
			
            // 类似，道具消耗
            if !state.Bag.IsExist(need_produce_items) {
                // 处理错误
            }

            changeItems := state.Bag.ProduceItems(need_produce_items)

			return nil
		},
	}
}

```

<div style="display: flex; align-items: center; margin: 1em 0;">
  <div style="flex-grow: 1; height: 1px; background-color: #ccc;"></div>
  <div style="margin: 0 10px; font-weight: bold; color: #666;">进阶</div>
  <div style="flex-grow: 1; height: 1px; background-color: #ccc;"></div>
</div>

> 通常 actor 的 state 只用于自身处理 events 时的修改， 但也有一些需求需要提供外部访问甚至需要多个 actor state 的同步操作

</br>

### 提供给外部访问和修改 (伪代码

```go
func MKCheckRoomState(ctx core.ActorContext) core.IChain {

    return &actor.DefaultChain{
        Handler: func(mw *router.MsgWrapper) error {

            state := ctx.GetValue(RoomStateType{}).(*room.RoomWrapper)

            mw.req.opt.WithValue(user.ROOM_ID, "new_value")

            // 使用阻塞操作
            // 通过 room 的 state 获取到目标 actor 的 id
            // 执行 check_and_update 语义，去修改其他 actor 的 state
            err := ctx.Call(router.Target{ID: state.UserID, Ev: events.ChangeUserState }, mw)
            if err != nil {
                // 如果执行失败，进行失败处理
            }

			return nil
		},
    }
}

func MKChangeUserState(ctx core.ActorContext) core.IChain {

    return &actor.DefaultChain{
        Handler: func(mw *router.MsgWrapper) error {

            state := ctx.GetValue(UserStateType{}).(*user.EntityWrapper)

            switch mw.req.opt.key {
                case user.ROOM_ID:
                    // 1.检查是否符合修改需求，比如通过房间状态等信息
                    if !checked { return err } // 通知调用者修改失败

                    // 2. 通过 state 的接口进行修改
                    state.Room.Change()
            }

			return nil
		},
    }
}

```

</br>

### 需要同步修改多个 actor 的 state
跳转到 [分布式事务(TCC)](zh-cn/pages-transaction.md) 页查看