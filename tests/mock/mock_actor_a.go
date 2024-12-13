package mock

import (
	"context"
	"fmt"
	"sync/atomic"

	"github.com/google/uuid"
	"github.com/pojol/braid/core"
	"github.com/pojol/braid/core/actor"
	"github.com/pojol/braid/def"
	"github.com/pojol/braid/lib/log"
	"github.com/pojol/braid/router/msg"
)

type mockActorA struct {
	*actor.Runtime
}

var RecenterCalcValue int32

func newMockA(p core.IActorBuilder) core.IActor {
	return &mockActorA{
		Runtime: &actor.Runtime{Id: p.GetID(), Ty: p.GetType(), Sys: p.GetSystem()},
	}
}

func (ra *mockActorA) Init(ctx context.Context) {
	ra.Runtime.Init(ctx)

	ra.RegisterEvent("reenter", func(ctx core.ActorContext) core.IChain {
		return &actor.DefaultChain{
			Handler: func(w *msg.Wrapper) error {

				calculateVal := 2
				w.ToBuilder().WithReqCustomFields(msg.Attr{Key: "calculateVal", Value: calculateVal})

				future := ctx.ReenterCall("mockb", "mockb", "clac", w)
				future.Then(func(w *msg.Wrapper) {
					val := msg.GetResCustomField[int](w, "calculateVal")
					atomic.CompareAndSwapInt32(&RecenterCalcValue, 0, int32(val*2))
				})

				return nil
			},
		}
	})

	ra.RegisterEvent("timeout", func(ctx core.ActorContext) core.IChain {
		return &actor.DefaultChain{
			Handler: func(w *msg.Wrapper) error {

				future := ctx.ReenterCall("mockb", "mockb", "timeout", w)
				future.Then(func(fw *msg.Wrapper) {

					if fw.Err != nil {
						w.Err = fw.Err
					}
				})

				return nil
			},
		}
	})

	ra.SubscriptionEvent(ra.Id, "offline_msg", func(ctx core.ActorContext) core.IChain {
		return &actor.DefaultChain{
			Handler: func(w *msg.Wrapper) error {
				log.InfoF("recv offline_msg %s", string(w.Req.Body))

				return nil
			},
		}
	})

	ra.RegisterEvent("chain", func(ctx core.ActorContext) core.IChain {
		return &actor.DefaultChain{
			Handler: func(w *msg.Wrapper) error {

				calculateVal := 2
				w.ToBuilder().WithReqCustomFields(msg.Attr{Key: "calculateVal", Value: calculateVal})

				future := ctx.ReenterCall("mockb", "mockb", "clac", w)
				future.Then(func(w *msg.Wrapper) {
					val := msg.GetResCustomField[int](w, "calculateVal")
					atomic.CompareAndSwapInt32(&RecenterCalcValue, 0, int32(val*2))
				}).Then(func(w *msg.Wrapper) {
					atomic.AddInt32(&RecenterCalcValue, 10)
				})

				return nil
			},
		}
	})

	ra.RegisterEvent("test_block", func(ctx core.ActorContext) core.IChain {
		return &actor.DefaultChain{
			Handler: func(w *msg.Wrapper) error {

				val := msg.GetReqCustomField[int](w, "randvalue")
				w.ToBuilder().WithReqCustomFields(msg.Attr{Key: "randvalue", Value: val + 1})
				ctx.Call("mockb", "mockb", "test_block", w)

				return nil
			},
		}
	})

	ra.RegisterEvent("tcc_succ", func(ctx core.ActorContext) core.IChain {
		return &actor.DefaultChain{
			Handler: func(w *msg.Wrapper) error {

				transactionID := uuid.New().String()

				bmsg := msg.NewBuilder(w.Ctx).WithReqCustomFields(def.TransactionID(transactionID))
				cmsg := msg.NewBuilder(w.Ctx).WithReqCustomFields(def.TransactionID(transactionID))

				bsucc := ctx.Call("mockb", "mockb", "tcc_succ", bmsg.Build())
				csucc := ctx.Call("mockc", "mockc", "tcc_succ", cmsg.Build())

				var err error

				if bsucc == nil && csucc == nil { // succ

					bconfirmmsg := msg.NewBuilder(w.Ctx).WithReqCustomFields(def.TransactionID(transactionID)).Build()
					err = ctx.Call("mockb", "mockb", "tcc_confirm", bconfirmmsg)
					if err != nil {
						/*
							err = ctx.Pub("mockb_tcc_confirm", bconfirmmsg.Req)
							if err != nil {
								fmt.Println("???")
							}
						*/
					}

					cconfirmmsg := msg.NewBuilder(w.Ctx).WithReqCustomFields(def.TransactionID(transactionID)).Build()
					err = ctx.Call("mockc", "mockc", "tcc_confirm", cconfirmmsg)
					if err != nil {
						/*
							err = ctx.Pub("mockc_tcc_confirm", cconfirmmsg.Req)
							if err != nil {
								fmt.Println("???")
							}
						*/
					}
				} else {
					fmt.Println("tcc call err", "b", bsucc, "c", csucc)
				}

				fmt.Println("mock a tcc_succ end")
				return nil
			},
		}
	})
}
