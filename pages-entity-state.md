# introduce state to actor


</br>

* [Binding State](#binding-state)
* [Setting and Getting State](#setting-and-getting-state)
* [Operating Through State](#operating-through-state)
---
* [Providing External Access and Modification](#providing-external-access-and-modification-pseudocode)
* [Synchronously Modifying Multiple Actor States](#synchronously-modifying-multiple-actor-states)

</br>

### Binding State

```go
type UserActor struct {
	*actor.Runtime
	state    *user.EntityWrapper    // 1. Bind state to actor
}

func NewUserActor(p core.IActorBuilder) core.IActor {
	return &UserActor{
		Runtime:   &actor.Runtime{Id: p.GetID(), Ty: p.GetType(), Sys: p.GetSystem()},
		state:    user.NewEntityWapper(p.GetID()),  // 2. Construct state
	}
}

func (a *UserActor) Init(ctx context.Context) {
    // 3. Load state
    err := a.state.Load(context.TODO())
	if err != nil {
		panic(fmt.Errorf("load user actor err %v", err.Error()))
	}
}

```

</br>

### Setting and Getting State

```go
func (a *UserActor) Init(ctx context.Context) {

    // 2. Fill state into context for easy access in events
	a.Context().WithValue(events.UserStateType{}, a.entity) 
    
}

// events_user_base.go
type UserStateType struct{} // 1. Define state type in context

func MkOperatorXXX(ctx core.ActorContext) core.IChain {
    return &actor.DefaultChain{
		Handler: func(mw *msg.Wrapper) error {

            // 3. Extract state for various modifications and calculations
			state := ctx.GetValue(UserStateType{}).(*user.EntityWrapper)
        }
    }
}

```

</br>

### Operating Through State

```go
func MkOperatorXXX(ctx core.ActorContext) core.IChain {
    return &actor.DefaultChain{
		Handler: func(mw *msg.Wrapper) error {

            // Extract state for various modifications and calculations
			state := ctx.GetValue(UserStateType{}).(*user.EntityWrapper)
			
            // For example, item consumption
            if !state.Bag.IsExist(need_produce_items) {
                // Handle error
            }

            changeItems := state.Bag.ProduceItems(need_produce_items)

			return nil
		},
	}
}

```

<div style="display: flex; align-items: center; margin: 1em 0;">
  <div style="flex-grow: 1; height: 1px; background-color: #ccc;"></div>
  <div style="margin: 0 10px; font-weight: bold; color: #666;">Advanced</div>
  <div style="flex-grow: 1; height: 1px; background-color: #ccc;"></div>
</div>

> Typically, an actor's state is only used for modifications when handling its own events, but there are some requirements that need external access or even synchronous operations on multiple actor states.

</br>

### Providing External Access and Modification (Pseudocode)

```go
func MKCheckRoomState(ctx core.ActorContext) core.IChain {

    return &actor.DefaultChain{
        Handler: func(mw *msg.Wrapper) error {

            state := ctx.GetValue(RoomStateType{}).(*room.RoomWrapper)

            mw.ToBuilder().WithReqCustomFields(fields.ROOM_ID(new_value))

            // Use blocking operation
            // Get the target actor's id through the room's state
            // Execute check_and_update semantics to modify other actor's state
            err := ctx.Call(router.Target{ID: state.UserID, Ev: events.ChangeUserState }, mw)
            if err != nil {
                // Handle failure if execution fails
            }

			return nil
		},
    }
}

func MKChangeUserState(ctx core.ActorContext) core.IChain {

    return &actor.DefaultChain{
        Handler: func(mw *msg.Wrapper) error {

            state := ctx.GetValue(UserStateType{}).(*user.EntityWrapper)

            switch mw.req.opt.key {
                case user.ROOM_ID:
                    // 1. Check if it meets modification requirements, e.g., through room status information
                    if !checked { return err } // Notify caller of modification failure

                    // 2. Modify through state interface
                    state.Room.Change()
            }

			return nil
		},
    }
}

```

</br>

### Synchronously Modifying Multiple Actor States
Jump to [Distributed Transactions (TCC)](pages-transaction.md) page to view