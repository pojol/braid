package log

import (
	"fmt"

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

type Helper interface {
	SystemReqInfo(id LogAccID, path LogPath, token LogToken, ip LogRealIP, format string, v ...interface{})
	SystemLog(level zapcore.Level, code LogECode, id LogAccID, path LogPath, token LogToken, ip LogRealIP, format string, v ...interface{})
	Log(level zapcore.Level, format string, v ...interface{})
	Sync()
}

type helper struct {
	logger *zap.Logger
	tag    string
}

func NewHelper(logger *zap.Logger, tag string) Helper {
	return &helper{logger: logger, tag: tag}
}

func (h *helper) SystemLog(level zapcore.Level, code LogECode, id LogAccID, path LogPath, token LogToken, ip LogRealIP, format string, v ...interface{}) {
	if !h.logger.Core().Enabled(level) {
		return
	}

	msg := fmt.Sprintf(format, v...)
	fields := h.getSystemFields(code, id, path, token, ip)
	switch level {
	case zapcore.DebugLevel:
		h.logger.Debug(msg, fields...)
	case zapcore.InfoLevel:
		h.logger.Info(msg, fields...)
	case zapcore.WarnLevel:
		h.logger.Warn(msg, fields...)
	case zapcore.ErrorLevel:
		h.logger.Error(msg, fields...)
	}
}

func (h *helper) SystemReqInfo(id LogAccID, path LogPath, token LogToken, ip LogRealIP, format string, v ...interface{}) {
	if !h.logger.Core().Enabled(zapcore.InfoLevel) {
		return
	}
	msg := fmt.Sprintf(format, v...)
	h.logger.Info(msg, zap.String(ServerTagKey, h.tag), id.field(), path.field(), token.field(), ip.field())
}

func (h *helper) Log(level zapcore.Level, format string, v ...interface{}) {
	if !h.logger.Core().Enabled(level) {
		return
	}

	msg := fmt.Sprintf(format, v...)
	field := zap.String(ServerTagKey, h.tag)
	switch level {
	case zapcore.DebugLevel:
		h.logger.Debug(msg, field)
	case zapcore.InfoLevel:
		h.logger.Info(msg, field)
	case zapcore.WarnLevel:
		h.logger.Warn(msg, field)
	case zapcore.ErrorLevel:
		h.logger.Error(msg, field)
	case zapcore.DPanicLevel:
		h.logger.DPanic(msg, field)
	case zapcore.PanicLevel:
		h.logger.Panic(msg, field)
	case zapcore.FatalLevel:
		h.logger.Fatal(msg, field)
	}
}

func (h *helper) Sync() {
	h.logger.Sync()
}

func (h *helper) getSystemFields(code LogECode, id LogAccID, path LogPath, token LogToken, ip LogRealIP) []zap.Field {
	return []zap.Field{
		zap.String(ServerTagKey, h.tag),
		id.field(),
		code.field(),
		path.field(),
		token.field(),
		ip.field(),
	}
}

func SystemDebug(code LogECode, id LogAccID, path LogPath, token LogToken, ip LogRealIP, format string, v ...interface{}) {
	sLog.SystemLog(zapcore.DebugLevel, code, id, path, token, ip, format, v...)
}

func SystemInfo(code LogECode, id LogAccID, path LogPath, token LogToken, ip LogRealIP, format string, v ...interface{}) {
	sLog.SystemLog(zapcore.InfoLevel, code, id, path, token, ip, format, v...)
}

func SystemWarn(code LogECode, id LogAccID, path LogPath, token LogToken, ip LogRealIP, format string, v ...interface{}) {
	sLog.SystemLog(zapcore.WarnLevel, code, id, path, token, ip, format, v...)
}

func SystemError(code LogECode, id LogAccID, path LogPath, token LogToken, ip LogRealIP, format string, v ...interface{}) {
	sLog.SystemLog(zapcore.ErrorLevel, code, id, path, token, ip, format, v...)
}

func SystemReqInfo(id LogAccID, path LogPath, token LogToken, ip LogRealIP, format string, v ...interface{}) {
	sLog.SystemReqInfo(id, path, token, ip, format, v...)
}

func DebugF(format string, v ...interface{}) {
	sLog.Log(zapcore.DebugLevel, format, v...)
}

func InfoF(format string, v ...interface{}) {
	sLog.Log(zapcore.InfoLevel, format, v...)
}

func ErrorF(format string, v ...interface{}) {
	sLog.Log(zapcore.ErrorLevel, format, v...)
}

func WarnF(format string, v ...interface{}) {
	sLog.Log(zapcore.WarnLevel, format, v...)
}

func PanicF(format string, v ...interface{}) {
	sLog.Log(zapcore.PanicLevel, format, v...)
}
