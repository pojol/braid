package msg

import (
	"context"

	"github.com/google/uuid"
	"github.com/pojol/braid/lib/warpwaitgroup"
	"github.com/pojol/braid/router"
)

// Parm nsq config
type WrapperParm struct {
	CustomObjSerialize ICustomSerialize // default msg pack
	CustomMapSerialize ICustomSerialize // default json
}

// Option config wraps
type WrapperOption func(*WrapperParm)

type Wrapper struct {
	Req *router.Message // The proto-defined Message
	Res *router.Message
	Ctx context.Context
	Err error

	parm WrapperParm
	Done chan struct{} // Used for synchronization
}

// NewMessage create new message
func newMessage(uid string) *router.Message {
	m := &router.Message{
		Header: &router.Header{
			ID: uid,
		},
	}
	return m
}

type WaitGroupKey struct{}

// MsgWrapperBuilder used to build MsgWrapper
type MsgBuilder struct {
	wrapper *Wrapper
}

func NewBuilder(ctx context.Context, opts ...WrapperOption) *MsgBuilder {
	uid := uuid.NewString()

	parm := WrapperParm{
		CustomMapSerialize: &CustomJsonSerialize{},
		CustomObjSerialize: &CustomObjectSerialize{},
	}

	if wc, ok := ctx.Value(WaitGroupKey{}).(*warpwaitgroup.WrapWaitGroup); ok {
		ctx = context.WithValue(ctx, WaitGroupKey{}, wc)
	} else {
		ctx = context.WithValue(ctx, WaitGroupKey{}, &warpwaitgroup.WrapWaitGroup{})
	}

	return &MsgBuilder{
		wrapper: &Wrapper{
			parm: parm,
			Ctx:  ctx,
			Req:  newMessage(uid),
			Res:  newMessage(uid),
		},
	}
}

func Swap(mw *Wrapper) *Wrapper {

	ctx := context.WithValue(mw.Ctx, WaitGroupKey{}, &warpwaitgroup.WrapWaitGroup{})

	return &Wrapper{
		Ctx: ctx,
		// 交换 Req 和 Res
		Req:  mw.Req,
		Res:  mw.Res,
		Done: make(chan struct{}),
	}
}

func (b *MsgBuilder) WithReqHeader(h *router.Header) *MsgBuilder {
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
		b.wrapper.Req.Header.Custom = h.Custom
	} else {
		// If either header is nil, directly set the header
		b.wrapper.Req.Header = h
	}
	return b
}

func (b *MsgBuilder) WithReqBody(byt []byte) *MsgBuilder {
	b.wrapper.Req.Body = byt
	return b
}

// WithRes set res header
func (b *MsgBuilder) WithResHeader(h *router.Header) *MsgBuilder {
	b.wrapper.Res.Header = h
	return b
}

func (b *MsgBuilder) WithResBody(byt []byte) *MsgBuilder {
	b.wrapper.Res.Body = byt
	return b
}

// Build build msg wrapper
func (b *MsgBuilder) Build() *Wrapper {
	return b.wrapper
}

func (mw *Wrapper) ToBuilder() *MsgBuilder {
	if mw == nil {
		return NewBuilder(context.Background())
	}

	return &MsgBuilder{
		wrapper: mw,
	}
}

////////

func (mw *Wrapper) GetWg() *warpwaitgroup.WrapWaitGroup {
	if wc, ok := mw.Ctx.Value(WaitGroupKey{}).(*warpwaitgroup.WrapWaitGroup); ok {
		return wc
	}
	return nil
}
