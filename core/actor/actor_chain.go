package actor

import (
	"github.com/pojol/braid/router"
)

type DefaultChain struct {
	Before  []EventHandler
	After   []EventHandler
	Handler EventHandler
}

func (c *DefaultChain) Execute(m *router.MsgWrapper) error {
	var err error

	for _, before := range c.Before {
		err = before(m)
		if err != nil {
			goto ext
		}
	}

	err = c.Handler(m)
	if err != nil {
		goto ext
	}

	for _, after := range c.After {
		err = after(m)
		if err != nil {
			goto ext
		}
	}

ext:
	return err
}
