package log

import (
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

var gBehaviorLog *zap.Logger

// ExportBehaviorLog 设置global日志(注不能在多线程环境下使用
func ExportBehaviorLog(log *zap.Logger) {
	if log != nil {
		gBehaviorLog = log
	} else {
		panic("logger point is nil!")
	}
}

// 调试日志
func InfoF(args ...zapcore.Field) {
	gBehaviorLog.Info("", args...)
}

// 警告类日志
func BSystemWarm(module, fc, msg string) {
	gBehaviorLog.Info("",
		zap.String("行为", "警告"),
		zap.String("module", module),
		zap.String("func", fc),
		zap.String("msg", msg),
	)
}

// 调试类日志
func BSystemDebug(module, fc, msg string) {
	gBehaviorLog.Info("",
		zap.String("行为", "调试"),
		zap.String("module", module),
		zap.String("func", fc),
		zap.String("msg", msg),
	)
}

// 错误类日志 (需要告警
func BSystemErr(module, fc, stack string) {
	gBehaviorLog.Info("",
		zap.String("行为", "错误"),
		zap.String("module", module),
		zap.String("func", fc),
		zap.String("stack", stack),
	)
}

// ServerLog -------------ServerLog-------------->>
type ServerLog struct {
	Helper
}

var sLog, _ = NewServerLogger("")

const (
	ServerTagKey = "ServerTag"
	PathKey      = "path"
	AccIDKey     = "accID"
	ECodeKey     = "eCode"
	TokenKey     = "token"
	RealIPKey    = "realIP"
)

func ExportSLog(log *zap.Logger, tag string) {
	if log != nil {
		sLog.Helper = NewHelper(log, tag)
	} else {
		panic("logger point is nil!")
	}
}

type LogECode int32

func (e LogECode) field() zap.Field {
	return zap.Int32(ECodeKey, int32(e))
}

type LogAccID string

func (s LogAccID) field() zap.Field {
	return zap.String(AccIDKey, string(s))
}

type LogPath string

func (s LogPath) field() zap.Field {
	return zap.String(PathKey, string(s))
}

type LogToken string

func (s LogToken) field() zap.Field {
	return zap.String(TokenKey, string(s))
}

type LogRealIP string

func (s LogRealIP) field() zap.Field {
	return zap.String(RealIPKey, string(s))
}

// <<-------------ServerLog--------------
