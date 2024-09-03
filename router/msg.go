package router

import "github.com/pojol/braid/lib/warpwaitgroup"

type MsgWrapper struct {
	Req    *Message // The proto-defined Message
	Res    *Message
	Entity interface{} // player, guild, social, chat ... object

	Wg   warpwaitgroup.WrapWaitGroup
	Done chan struct{} // Used for synchronization
}
