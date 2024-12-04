package mock

import (
	"context"
	"fmt"
	"time"

	"github.com/pojol/braid/core"
	"github.com/pojol/braid/core/actor"
	"github.com/pojol/braid/def"
	"github.com/pojol/braid/router/msg"
)

var MockCTccValue = 22

type mockActorC struct {
	*actor.Runtime
	tcc *TCC
}

func newMockC(p core.IActorBuilder) core.IActor {
	return &mockActorC{
		Runtime: &actor.Runtime{Id: p.GetID(), Ty: p.GetType(), Sys: p.GetSystem()},
		tcc:     &TCC{stateMap: make(map[string]*tccState)},
	}
}

func (a *mockActorC) Init(ctx context.Context) {
	a.Runtime.Init(ctx)

	a.RegisterEvent("ping", func(ctx core.ActorContext) core.IChain {
		return &actor.DefaultChain{
			Handler: func(w *msg.Wrapper) error {
				w.ToBuilder().WithResCustomFields(msg.Attr{Key: "pong", Value: "pong"})
				return nil
			},
		}
	})

	a.RegisterEvent("test_block", func(ctx core.ActorContext) core.IChain {
		return &actor.DefaultChain{
			Handler: func(w *msg.Wrapper) error {

				val := msg.GetReqField[int](w, "randvalue")
				w.ToBuilder().WithResCustomFields(msg.Attr{Key: "randvalue", Value: val + 1})

				return nil
			},
		}
	})

	a.RegisterEvent("tcc_succ", func(ctx core.ActorContext) core.IChain {
		return &actor.DefaultChain{
			Handler: func(w *msg.Wrapper) error {

				transID := msg.GetReqField[string](w, def.KeyTranscationID)

				a.tcc.stateMap[transID] = &tccState{
					originValue:  MockCTccValue,
					currentValue: 222,
					status:       "try",
					createdAt:    time.Now(),
				}

				MockCTccValue = 222
				fmt.Println("succ mock c value", MockCTccValue)
				return nil
			},
		}
	})

	a.RegisterEvent("tcc_confirm", func(ctx core.ActorContext) core.IChain {
		return &actor.DefaultChain{
			Handler: func(w *msg.Wrapper) error {
				transID := msg.GetReqField[string](w, def.KeyTranscationID)

				if state, exists := a.tcc.stateMap[transID]; exists {
					state.status = "confirmed"
					delete(a.tcc.stateMap, transID)
					return nil
				}
				return fmt.Errorf("transaction %s not found", transID)
			},
		}
	})

	a.RegisterEvent("tcc_cancel", func(ctx core.ActorContext) core.IChain {
		return &actor.DefaultChain{
			Handler: func(w *msg.Wrapper) error {
				transID := msg.GetReqField[string](w, def.KeyTranscationID)

				if state, exists := a.tcc.stateMap[transID]; exists {
					MockCTccValue = state.originValue
					state.status = "cancelled"
					delete(a.tcc.stateMap, transID)
					return nil
				}
				return fmt.Errorf("transaction %s not found", transID)
			},
		}
	})
}
