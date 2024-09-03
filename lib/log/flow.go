package log

import (
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

// 基础属性字段,每条行为日志都打印相关信息
type CommonFlow struct {
	Fields []zap.Field
}

type BaseFlow struct {
	*CommonFlow
	BaseFields []zap.Field
}

func (f *BaseFlow) SetCommon(cf *CommonFlow) {
	f.CommonFlow = cf
}

func (f *BaseFlow) ZapFields() []zap.Field {
	var cField []zap.Field
	if f.CommonFlow != nil {
		for _, v := range f.CommonFlow.Fields {
			cField = append(cField, v)
		}
	}

	var pField []zap.Field
	for _, v := range f.BaseFields {
		pField = append(pField, v)
	}
	return append(cField, pField...)
}

func (f *BaseFlow) TaDataFields(fieldKey string, isSubObject func(k string) bool) []zap.Field {
	enc := zapcore.NewMapObjectEncoder()
	fields := f.ZapFields()
	if len(fieldKey) <= 0 {
		return fields
	}
	var ret []zap.Field
	hasSubObject := false
	for _, v := range fields {
		if isSubObject(v.Key) {
			hasSubObject = true
			v.AddTo(enc)
		} else {
			ret = append(ret, v)
		}
	}
	if hasSubObject {
		ret = append(ret, zap.Any(fieldKey, enc.Fields))
	}

	return ret
}
