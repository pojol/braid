package router

import (
	"sync"

	"github.com/pojol/braid/lib/warpwaitgroup"
)

type MsgWrapper struct {
	Req    *Message // The proto-defined Message
	Res    *Message
	Entity interface{} // player, guild, social, chat ... object

	Wg   warpwaitgroup.WrapWaitGroup
	Done chan struct{} // Used for synchronization
}

func GetMsg() *MsgWrapper {
	return &MsgWrapper{
		Req: &Message{
			Header: &Header{
				Custom: make(map[string]string),
			},
		},
		Res: &Message{
			Header: &Header{
				Custom: make(map[string]string),
			},
		},
	}
}

var msgPool = sync.Pool{
	New: func() interface{} {
		return &MsgWrapper{
			Req: &Message{
				Header: &Header{},
			},
			Res: &Message{
				Header: &Header{
					Custom: make(map[string]string),
				},
			},
		}
	},
}

// GetMsg retrieves a MsgWrapper from the pool
func GetMsgWithPool() *MsgWrapper {
	return msgPool.Get().(*MsgWrapper)
}

// PutMsg returns a MsgWrapper to the pool
func PutMsg(msg *MsgWrapper) {
	// Clear the message before returning it to the pool
	msg.Req.Header.Reset()
	msg.Res.Header.Reset()
	for k := range msg.Res.Header.Custom {
		delete(msg.Res.Header.Custom, k)
	}
	msgPool.Put(msg)
}
