package msg

import (
	"encoding/json"
	fmt "fmt"

	"github.com/pojol/braid/lib/log"
)

func (b *MsgBuilder) WithReqCustomFields(attrs ...Attr) *MsgBuilder {

	data := make(map[string]any, len(attrs))

	/*
	   Optimization suggestion:
	   - Wrap request into: wrapreq { *req, *customptr }
	   - Postpone custom field serialization until just before RPC call
	   - Benefits: Eliminates unnecessary serialization overhead for intra-node actor communications
	*/
	if len(b.wrapper.Req.Header.Custom) > 0 {
		if err := b.wrapper.parm.CustomMapSerialize.Decode(b.wrapper.Req.Header.Custom, &data); err != nil {
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

	byt, err := b.wrapper.parm.CustomMapSerialize.Encode(data)
	if err != nil {
		b.wrapper.Err = fmt.Errorf("marshal request body failed: %w", err)
		return b
	}
	b.wrapper.Req.Header.Custom = byt

	return b
}

func (b *MsgBuilder) WithResCustomFields(attrs ...Attr) *MsgBuilder {
	data := make(map[string]any, len(attrs))

	if len(b.wrapper.Res.Header.Custom) > 0 {
		if err := b.wrapper.parm.CustomMapSerialize.Decode(b.wrapper.Res.Header.Custom, &data); err != nil {
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

	byt, err := b.wrapper.parm.CustomMapSerialize.Encode(data)
	if err != nil {
		b.wrapper.Err = fmt.Errorf("marshal request body failed: %w", err)
		return b
	}
	b.wrapper.Res.Header.Custom = byt

	return b
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

func GetReqCustomField[T any](msg *Wrapper, key string) T {
	var zero T

	data, err := msg.GetReqCustomMap()
	if err != nil {
		log.WarnF("[braid.router] get req body map err %v", err.Error())
		return zero
	}

	val, ok := data[key]
	if !ok {
		log.InfoF("[braid.router] key %q not found in request body", key)
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

func GetResCustomField[T any](msg *Wrapper, key string) T {
	var zero T

	data, err := msg.GetResCustomMap()
	if err != nil {
		log.WarnF("[braid.router] get res body map err %v", err.Error())
		return zero
	}

	val, ok := data[key]
	if !ok {
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
