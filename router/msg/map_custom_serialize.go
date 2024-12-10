package msg

import (
	"encoding/json"

	"github.com/vmihailenco/msgpack/v5"
)

type ICustomSerialize interface {
	Decode(data []byte, v any) error
	Encode(v any) ([]byte, error)
}

type CustomJsonSerialize struct {
}

func (c *CustomJsonSerialize) Decode(data []byte, v any) error {
	return json.Unmarshal(data, v)
}

func (c *CustomJsonSerialize) Encode(v any) ([]byte, error) {
	return json.Marshal(v)
}

type CustomObjectSerialize struct {
}

func (c *CustomObjectSerialize) Decode(data []byte, v any) error {
	return msgpack.Unmarshal(data, v)
}

func (c *CustomObjectSerialize) Encode(v any) ([]byte, error) {
	return msgpack.Marshal(v)
}
