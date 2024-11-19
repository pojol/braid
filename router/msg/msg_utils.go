package msg

import (
	"encoding/json"
	fmt "fmt"
	"reflect"
	"time"
)

type Attr struct {
	Key   string
	Value any
}

func isJSONSerializable(v interface{}) error {
	switch val := v.(type) {
	case nil, string, bool, int, int8, int16, int32, int64,
		uint, uint8, uint16, uint32, uint64, float32, float64:
		return nil

	case time.Time:
		return nil

	case json.Marshaler:
		return nil

	case []interface{}:
		for _, item := range val {
			if err := isJSONSerializable(item); err != nil {
				return fmt.Errorf("invalid slice element: %w", err)
			}
		}
		return nil

	case map[string]interface{}:
		for k, v := range val {
			if err := isJSONSerializable(v); err != nil {
				return fmt.Errorf("invalid value for key '%s': %w", k, err)
			}
		}
		return nil

	default:
		// 检查是否为切片或数组
		if rt := reflect.TypeOf(v); rt != nil && rt.Kind() == reflect.Slice {
			rv := reflect.ValueOf(v)
			for i := 0; i < rv.Len(); i++ {
				if err := isJSONSerializable(rv.Index(i).Interface()); err != nil {
					return fmt.Errorf("invalid slice element at index %d: %w", i, err)
				}
			}
			return nil
		}

		return fmt.Errorf("unsupported type %T", v)
	}
}
