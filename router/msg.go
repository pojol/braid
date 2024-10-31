package router

import (
	"context"

	"github.com/google/uuid"
	"github.com/pojol/braid/def"
	"github.com/pojol/braid/lib/log"
	"github.com/pojol/braid/lib/warpwaitgroup"
)

type MsgWrapper struct {
	Req *Message // The proto-defined Message
	Res *Message
	Ctx context.Context
	Err error

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

func NewMsgWrap(ctx context.Context) *MsgWrapperBuilder {
	uid := uuid.NewString()
	return &MsgWrapperBuilder{
		wrapper: MsgWrapper{
			Ctx: ctx,
			Req: newMessage(uid),
			Res: newMessage(uid),
		},
	}
}

func (b *MsgWrapperBuilder) WithReqHeader(h *Header) *MsgWrapperBuilder {
	if b.wrapper.Req.Header != nil && h != nil {
		// Copy fields from the input header to existing header
		b.wrapper.Req.Header.ID = h.ID
		b.wrapper.Req.Header.Event = h.Event
		b.wrapper.Req.Header.OrgActorID = h.OrgActorID
		b.wrapper.Req.Header.OrgActorType = h.OrgActorType
		b.wrapper.Req.Header.Token = h.Token
		b.wrapper.Req.Header.PrevActorType = h.PrevActorType
		b.wrapper.Req.Header.TargetActorID = h.TargetActorID
		b.wrapper.Req.Header.TargetActorType = h.TargetActorType

		// Deep copy the Custom map
		if h.Custom != nil {
			if b.wrapper.Req.Header.Custom == nil {
				b.wrapper.Req.Header.Custom = make(map[string]string)
			}
			for k, v := range h.Custom {
				b.wrapper.Req.Header.Custom[k] = v
			}
		}
	} else {
		// If either header is nil, directly set the header
		b.wrapper.Req.Header = h
	}
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

func (b *MsgWrapperBuilder) WithGateID(id string) *MsgWrapperBuilder {
	b.WithReqCustom(def.CustomGateID, id)
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

// Build build msg wrapper
func (b *MsgWrapperBuilder) Build() *MsgWrapper {
	return &b.wrapper
}

func (mw *MsgWrapper) GetGateID() string {
	if mw == nil {
		log.WarnF("message wrapper is nil")
		return ""
	}
	if mw.Req == nil || mw.Req.Header == nil {
		log.WarnF("invalid message structure: missing header")
		return ""
	}
	if mw.Req.Header.Custom == nil {
		log.WarnF("custom field map is nil")
		return ""
	}

	gateID, exists := mw.Req.Header.Custom[def.CustomGateID]
	if !exists {
		log.WarnF("gate ID not found in message")
		return ""
	}
	if gateID == "" {
		log.WarnF("gate ID is empty")
		return ""
	}

	return gateID
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
