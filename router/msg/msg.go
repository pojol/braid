package msg

import (
	"context"
	"encoding/json"
	fmt "fmt"

	"github.com/google/uuid"
	"github.com/pojol/braid/lib/log"
	"github.com/pojol/braid/lib/warpwaitgroup"
	"github.com/pojol/braid/router"
)

type Wrapper struct {
	Req *router.Message // The proto-defined Message
	Res *router.Message
	Ctx context.Context
	Err error

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

func NewBuilder(ctx context.Context) *MsgBuilder {
	uid := uuid.NewString()

	if wc, ok := ctx.Value(WaitGroupKey{}).(*warpwaitgroup.WrapWaitGroup); ok {
		ctx = context.WithValue(ctx, WaitGroupKey{}, wc)
	} else {
		ctx = context.WithValue(ctx, WaitGroupKey{}, &warpwaitgroup.WrapWaitGroup{})
	}

	return &MsgBuilder{
		wrapper: &Wrapper{
			Ctx: ctx,
			Req: newMessage(uid),
			Res: newMessage(uid),
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

func (b *MsgBuilder) WithReqCustomFields(attrs ...Attr) *MsgBuilder {

	data := make(map[string]any, len(attrs))

	if len(b.wrapper.Req.Header.Custom) > 0 {
		if err := json.Unmarshal(b.wrapper.Req.Header.Custom, &data); err != nil {
			b.wrapper.Err = fmt.Errorf("unmarshal existing custom fields failed: %w", err)
			return b
		}
	}

	for _, attr := range attrs {
		data[attr.Key] = attr.Value
	}

	return b.WithReqCustomFieldsMap(data)
}

func (b *MsgBuilder) WithReqCustomFieldsMap(data map[string]any) *MsgBuilder {

	if err := isJSONSerializable(data); err != nil {
		b.wrapper.Err = fmt.Errorf("invalid data structure: %w", err)
		return b
	}

	byt, err := json.Marshal(data)
	if err != nil {
		b.wrapper.Err = fmt.Errorf("marshal request body failed: %w", err)
		return b
	}
	b.wrapper.Req.Header.Custom = byt

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

func (b *MsgBuilder) WithResCustomFields(attrs ...Attr) *MsgBuilder {
	data := make(map[string]any, len(attrs))

	if len(b.wrapper.Res.Header.Custom) > 0 {
		if err := json.Unmarshal(b.wrapper.Res.Header.Custom, &data); err != nil {
			b.wrapper.Err = fmt.Errorf("unmarshal existing custom fields failed: %w", err)
			return b
		}
	}

	for _, attr := range attrs {
		data[attr.Key] = attr.Value
	}

	return b.WithResCustomFieldsMap(data)
}

func (b *MsgBuilder) WithResCustomFieldsMap(data map[string]any) *MsgBuilder {

	if err := isJSONSerializable(data); err != nil {
		b.wrapper.Err = fmt.Errorf("invalid data structure: %w", err)
		return b
	}

	byt, err := json.Marshal(data)
	if err != nil {
		b.wrapper.Err = fmt.Errorf("marshal request body failed: %w", err)
		return b
	}
	b.wrapper.Res.Header.Custom = byt

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

// GetReqCustomMap gets the custom fields map from the message
//
// Note: The custom map is serialized using JSON, all value types need to be carefully converted
// (e.g., numeric types like int will be serialized to float64 and need to be converted back manually)
func (mw *Wrapper) GetReqCustomMap() (map[string]any, error) {
	if len(mw.Req.Header.Custom) == 0 {
		return nil, fmt.Errorf("empty request body")
	}
	var data map[string]any
	err := json.Unmarshal(mw.Req.Header.Custom, &data)
	return data, err
}

func GetReqField[T any](msg *Wrapper, key string) T {
	var zero T

	data, err := msg.GetReqCustomMap()
	if err != nil {
		log.WarnF("[braid.router] get req body map err %v", err.Error())
		return zero
	}

	val, ok := data[key]
	if !ok {
		log.WarnF("[braid.router] key %q not found in request body", key)
		return zero
	}

	// Type assert the value
	if typed, ok := val.(T); ok {
		return typed
	}

	// Type assert the value
	switch any(zero).(type) {
	case int:
		if f, ok := val.(float64); ok {
			return any(int(f)).(T)
		}
	case uint:
		if f, ok := val.(float64); ok {
			return any(uint(f)).(T)
		}
	case int16:
		if f, ok := val.(float64); ok {
			return any(int16(f)).(T)
		}
	case uint16:
		if f, ok := val.(float64); ok {
			return any(uint16(f)).(T)
		}
	case int32:
		if f, ok := val.(float64); ok {
			return any(int32(f)).(T)
		}
	case uint32:
		if f, ok := val.(float64); ok {
			return any(uint32(f)).(T)
		}
	case int64:
		if f, ok := val.(float64); ok {
			return any(int64(f)).(T)
		}
	case uint64:
		if f, ok := val.(float64); ok {
			return any(uint64(f)).(T)
		}
	case float32:
		if f, ok := val.(float64); ok {
			return any(float32(f)).(T)
		}
	}

	log.WarnF("[braid.router] type assertion failed for key %q: expected %T, got %T", key, zero, val)
	return zero
}

// GetResCustomMap gets the custom fields map from the message
//
// Note: The custom map is serialized using JSON, all value types need to be carefully converted
// (e.g., numeric types like int will be serialized to float64 and need to be converted back manually)
func (mw *Wrapper) GetResCustomMap() (map[string]any, error) {
	if len(mw.Res.Header.Custom) == 0 {
		return nil, fmt.Errorf("empty resuest body")
	}
	var data map[string]any
	err := json.Unmarshal(mw.Res.Header.Custom, &data)
	return data, err
}

func GetResField[T any](msg *Wrapper, key string) T {
	var zero T

	data, err := msg.GetResCustomMap()
	if err != nil {
		log.WarnF("[braid.router] get res body map err %v", err.Error())
		return zero
	}

	val, ok := data[key]
	if !ok {
		log.WarnF("[braid.router] key %q not found in response body", key)
		return zero
	}

	// Type assert the value
	if typed, ok := val.(T); ok {
		return typed
	}

	// Type assert the value
	switch any(zero).(type) {
	case int:
		if f, ok := val.(float64); ok {
			return any(int(f)).(T)
		}
	case uint:
		if f, ok := val.(float64); ok {
			return any(uint(f)).(T)
		}
	case int16:
		if f, ok := val.(float64); ok {
			return any(int16(f)).(T)
		}
	case uint16:
		if f, ok := val.(float64); ok {
			return any(uint16(f)).(T)
		}
	case int32:
		if f, ok := val.(float64); ok {
			return any(int32(f)).(T)
		}
	case uint32:
		if f, ok := val.(float64); ok {
			return any(uint32(f)).(T)
		}
	case int64:
		if f, ok := val.(float64); ok {
			return any(int64(f)).(T)
		}
	case uint64:
		if f, ok := val.(float64); ok {
			return any(uint64(f)).(T)
		}
	case float32:
		if f, ok := val.(float64); ok {
			return any(float32(f)).(T)
		}
	}

	log.WarnF("[braid.router] type assertion failed for key %q: expected %T, got %T", key, zero, val)
	return zero
}
