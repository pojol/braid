package mock

import (
	"context"
	"fmt"
	"sync/atomic"
	"time"

	"github.com/pojol/braid/core"
	"github.com/pojol/braid/core/actor"
	"github.com/pojol/braid/def"
	"github.com/pojol/braid/router/msg"
)

var MockBTccValue = 11
var BechmarkCallReceivedMessageCount int64

type mockActorB struct {
	*actor.Runtime
	tcc *TCC
}

func newMockB(p core.IActorBuilder) core.IActor {
	return &mockActorB{
		Runtime: &actor.Runtime{Id: p.GetID(), Ty: p.GetType(), Sys: p.GetSystem()},
		tcc:     &TCC{stateMap: make(map[string]*tccState)},
	}
}

func (a *mockActorB) Init(ctx context.Context) {
	a.Runtime.Init(ctx)

	a.OnEvent("clac", func(ctx core.ActorContext) core.IChain {
		return &actor.DefaultChain{
			Handler: func(w *msg.Wrapper) error {

				val := msg.GetReqCustomField[int](w, "calculateVal")
				w.ToBuilder().WithResCustomFields(msg.Attr{Key: "calculateVal", Value: val + 2})

				return nil
			},
		}
	})

	a.OnEvent("call_benchmark", func(ctx core.ActorContext) core.IChain {
		return &actor.DefaultChain{
			Handler: func(w *msg.Wrapper) error {
				atomic.AddInt64(&BechmarkCallReceivedMessageCount, 1)
				return nil
			},
		}
	})

	a.OnEvent("timeout", func(ctx core.ActorContext) core.IChain {
		return &actor.DefaultChain{
			Handler: func(w *msg.Wrapper) error {
				time.Sleep(time.Second * 5)
				return nil
			},
		}
	})

	a.OnEvent("test_block", func(ctx core.ActorContext) core.IChain {
		return &actor.DefaultChain{
			Handler: func(w *msg.Wrapper) error {

				val := msg.GetReqCustomField[int](w, "randvalue")
				w.ToBuilder().WithReqCustomFields(msg.Attr{Key: "randvalue", Value: val + 1})
				ctx.Call("mockc", "mockc", "test_block", w)

				return nil
			},
		}
	})

	a.OnEvent("tcc_succ", func(ctx core.ActorContext) core.IChain {
		return &actor.DefaultChain{
			Handler: func(w *msg.Wrapper) error {

				transID := msg.GetReqCustomField[string](w, def.KeyTranscationID)
				a.tcc.stateMap[transID] = &tccState{
					originValue:  MockBTccValue,
					currentValue: 111,
					status:       "try",
					createdAt:    time.Now(),
				}

				MockBTccValue = 111
				fmt.Println("succ mock b value", MockBTccValue)
				return nil
			},
		}
	})

	a.OnEvent("tcc_confirm", func(ctx core.ActorContext) core.IChain {
		return &actor.DefaultChain{
			Handler: func(w *msg.Wrapper) error {
				transID := msg.GetReqCustomField[string](w, def.KeyTranscationID)

				if state, exists := a.tcc.stateMap[transID]; exists {
					state.status = "confirmed"
					delete(a.tcc.stateMap, transID)
					return nil
				}
				return fmt.Errorf("transaction %s not found", transID)
			},
		}
	})

	a.OnEvent("tcc_cancel", func(ctx core.ActorContext) core.IChain {
		return &actor.DefaultChain{
			Handler: func(w *msg.Wrapper) error {
				transID := msg.GetReqCustomField[string](w, def.KeyTranscationID)

				if state, exists := a.tcc.stateMap[transID]; exists {
					MockBTccValue = state.originValue
					state.status = "cancelled"
					delete(a.tcc.stateMap, transID)
					return nil
				}
				return fmt.Errorf("transaction %s not found", transID)
			},
		}
	})
}
