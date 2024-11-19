package log

import (
	"fmt"
	"runtime"
	"strings"

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

func getStackTrace() string {
	// 分配调用栈空间
	const depth = 64
	var pcs [depth]uintptr
	n := runtime.Callers(2, pcs[:])
	frames := runtime.CallersFrames(pcs[:n])

	var builder strings.Builder
	builder.WriteString("\nStack Trace:\n")

	for {
		frame, more := frames.Next()
		builder.WriteString(fmt.Sprintf("%s\n\t%s:%d\n",
			frame.Function,
			frame.File,
			frame.Line))
		if !more {
			break
		}
	}
	return builder.String()
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

	var stack_trace zapcore.Field
	msg := fmt.Sprintf(format, v...)

	if level == zapcore.ErrorLevel || level == zapcore.WarnLevel {
		stackInfo := getStackTrace()
		stack_trace = zap.String("stack_trace", stackInfo)

		// 控制台输出，使用不同颜色区分
		if level == zapcore.ErrorLevel {
			fmt.Printf("\033[31m[ERROR] %s\n%s\033[0m\n", msg, stackInfo) // 红色
		} else {
			fmt.Printf("\033[33m[WARN] %s\n%s\033[0m\n", msg, stackInfo) // 黄色
		}
	}

	field := zap.String(ServerTagKey, h.tag)
	switch level {
	case zapcore.DebugLevel:
		h.logger.Debug(msg, field)
	case zapcore.InfoLevel:
		h.logger.Info(msg, field)
	case zapcore.WarnLevel:
		h.logger.Warn(msg, field, stack_trace)
	case zapcore.ErrorLevel:
		h.logger.Error(msg, field, stack_trace)
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
