package log

import (
	"fmt"

	"go.uber.org/zap"
)

type LogField interface {
	GetField() zap.Field
	GetKey() string
	GetType() string
}

type LogFieldType = string

var LogFieldTypeInt LogFieldType = "int32"
var LogFieldTypeInt64 LogFieldType = "int64"
var LogFieldTypeUint LogFieldType = "uint32"
var LogFieldTypeUint64 LogFieldType = "uint64"
var LogFieldTypeFloat64 LogFieldType = "float64"
var LogFieldTypeBool LogFieldType = "bool"
var LogFieldTypeString LogFieldType = "string"
var LogFieldTypeDate LogFieldType = "date"

var logFields []logField

type logField struct {
	Key   string
	Value LogFieldType
}

func GetLogFields() []logField {
	return logFields
}

func RegisterLogField(k string, v LogFieldType) {
	for _, field := range logFields {
		if field.Key == k {
			panic(fmt.Errorf("%v is repeated", k))
		}
	}

	logFields = append(logFields, logField{
		Key:   k,
		Value: v,
	})
}
