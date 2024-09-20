package router

import (
	"github.com/google/uuid"
	"github.com/pojol/braid/lib/warpwaitgroup"
)

type MsgWrapper struct {
	Req    *Message // The proto-defined Message
	Res    *Message
	Entity interface{} // player, guild, social, chat ... object

	Wg   warpwaitgroup.WrapWaitGroup
	Done chan struct{} // Used for synchronization
}

// NewMessage create new message
func newMessage(uid string) *Message {
	m := &Message{
		Header: &Header{
			ID:     uid,
			Custom: make(map[string]string),
		},
	}
	return m
}

// MsgWrapperBuilder used to build MsgWrapper
type MsgWrapperBuilder struct {
	wrapper MsgWrapper
}

func NewMsg() *MsgWrapperBuilder {
	uid := uuid.NewString()
	return &MsgWrapperBuilder{
		wrapper: MsgWrapper{
			Req: newMessage(uid),
			Res: newMessage(uid),
		},
	}
}

// WithReq set req header
func (b *MsgWrapperBuilder) WithReqHeader(h *Header) *MsgWrapperBuilder {
	b.wrapper.Req.Header = h
	return b
}

func (b *MsgWrapperBuilder) WithReqBody(byt []byte) *MsgWrapperBuilder {
	b.wrapper.Req.Body = byt
	return b
}

func (b *MsgWrapperBuilder) WithReqCustom(key, value string) *MsgWrapperBuilder {
	if b.wrapper.Req.Header.Custom == nil {
		b.wrapper.Req.Header.Custom = make(map[string]string)
	}
	b.wrapper.Req.Header.Custom[key] = value
	return b
}

// WithRes set res header
func (b *MsgWrapperBuilder) WithResHeader(h *Header) *MsgWrapperBuilder {
	b.wrapper.Res.Header = h
	return b
}

func (b *MsgWrapperBuilder) WithResBody(byt []byte) *MsgWrapperBuilder {
	b.wrapper.Res.Body = byt
	return b
}

func (b *MsgWrapperBuilder) WithResCustom(key, value string) *MsgWrapperBuilder {
	if b.wrapper.Res.Header.Custom == nil {
		b.wrapper.Res.Header.Custom = make(map[string]string)
	}
	b.wrapper.Res.Header.Custom[key] = value
	return b
}

// WithEntity set entity
func (b *MsgWrapperBuilder) WithEntity(entity interface{}) *MsgWrapperBuilder {
	b.wrapper.Entity = entity
	return b
}

// Build build msg wrapper
func (b *MsgWrapperBuilder) Build() *MsgWrapper {
	return &b.wrapper
}

/*
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
*/
