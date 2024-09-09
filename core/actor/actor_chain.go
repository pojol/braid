package actor

import (
	"context"

	"github.com/pojol/braid/router"
)

type DefaultChain struct {
	Before  []MiddlewareHandler
	After   []MiddlewareHandler
	Handler EventHandler
}

func (c *DefaultChain) Execute(ctx context.Context, m *router.MsgWrapper) error {
	var err error

	for _, before := range c.Before {
		err = before(ctx, m)
		if err != nil {
			goto ext
		}
	}

	err = c.Handler(ctx, m)
	if err != nil {
		goto ext
	}

	for _, after := range c.After {
		err = after(ctx, m)
		if err != nil {
			goto ext
		}
	}

ext:
	return err
}
