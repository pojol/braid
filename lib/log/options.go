package log

import (
	"fmt"
	"os"
	"path"
	"time"

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

const (
	DefaultMaxSize         = 1024
	DefaultMaxAge          = 7
	DefaultMaxBackups      = 30
	DefaultCompress        = false
	DefaultOutStd          = false
	DefaultTimestampFormat = "2006-01-02 15:04:05.000"
	DefaultLevel           = zapcore.InfoLevel

	// rotatelogs 默认参数
	RotateDefaultRotationSize  = 1 * 1024 * 1024 * 1024 // 1G
	RotateDefaultRotationTime  = 1 * time.Hour
	RotateDefaultMaxAge        = 0 // if both DefaultMaxAge and DefaultRotationCount are 0, give maxAge a sane default "7 * 24 * time.Hour"
	RotateDefaultRotationCount = 0 // 0 means this option is disabled.
)

type Options struct {
	Suffix      string                 // 日志文件后缀
	MaxSize     int                    // 每个日志文件保存的最大尺寸 单位：M
	MaxAge      int                    // 文件最多保存多少天
	MaxBackups  int                    // 日志文件最多保存多少个备份
	Compress    bool                   // 是否压缩
	OutStd      bool                   // 是否输出到控制台
	EncoderConf *zapcore.EncoderConfig // zap日志编码
	Level       zapcore.Level          // 日志等级
	Caller      bool                   // 是否打印行号,函数
	CallerFunc  bool                   // 是否打印函数
	CallerSkip  int                    // Caller skip frame count for caller info

	// rotatelogs使用参数
	// log filename pattern, if GlobPattern is empty, log will not be written to file
	GlobPattern string
	// sets the symbolic link name that gets linked to the current file name being used.
	LinkName string
	// (bytes) the log file size between rotation
	RotationSize int64
	// the time between rotation
	RotationTime time.Duration
	// the max age of a log file before it gets purged from the file system.
	// if both RotationCount and MaxAge are 0, give maxAge a sane default "7 * 24 * time.Hour"
	LogMaxAge time.Duration
	// the number of files should be kept before it gets purged from the file system
	// if both RotationCount and MaxAge are 0, give maxAge a sane default "7 * 24 * time.Hour"
	RotationCount uint
}

type Option func(*Options) error

func NewOptions(opts ...Option) (*Options, error) {
	opt := &Options{
		Suffix:     "", // 默认为空，不输出到文件
		MaxSize:    DefaultMaxSize,
		MaxAge:     DefaultMaxAge,
		MaxBackups: DefaultMaxBackups,
		Compress:   DefaultCompress,
		OutStd:     DefaultOutStd,
		EncoderConf: &zapcore.EncoderConfig{
			TimeKey:       "Time",
			LevelKey:      "Level",
			CallerKey:     "Caller",
			MessageKey:    "Msg",
			StacktraceKey: "StackTrace",
			LineEnding:    zapcore.DefaultLineEnding,
			EncodeLevel:   zapcore.CapitalLevelEncoder,
			EncodeTime:    zapcore.ISO8601TimeEncoder, // ISO8601 UTC 时间格式，filebeat格式要求
			//EncodeTime:    zapcore.TimeEncoderOfLayout(DefaultTimestampFormat),
			EncodeCaller: zapcore.ShortCallerEncoder,
		},
		Level:         DefaultLevel,
		Caller:        false,
		CallerFunc:    false,
		CallerSkip:    0,
		GlobPattern:   "", // 默认为空，不输出到文件
		LinkName:      "",
		RotationSize:  RotateDefaultRotationSize,
		RotationTime:  RotateDefaultRotationTime,
		LogMaxAge:     RotateDefaultMaxAge,
		RotationCount: RotateDefaultRotationCount,
	}

	for _, o := range opts {
		if err := o(opt); err != nil {
			return nil, err
		}
	}

	return opt, nil
}

func WithSuffix(s string) Option {
	return func(o *Options) error {
		o.Suffix = s
		return nil
	}
}

func WithMaxSize(s int) Option {
	return func(o *Options) error {
		o.MaxSize = s
		return nil
	}
}

func WithMaxBackups(b int) Option {
	return func(o *Options) error {
		o.MaxBackups = b
		return nil
	}
}

func WithCompress(c bool) Option {
	return func(o *Options) error {
		o.Compress = c
		return nil
	}
}

func WithOutStd(s bool) Option {
	return func(o *Options) error {
		o.OutStd = s
		return nil
	}
}

func WithEncoderConf(e *zapcore.EncoderConfig) Option {
	return func(o *Options) error {
		o.EncoderConf = e
		return nil
	}
}

func WithLevel(l zapcore.Level) Option {
	return func(o *Options) error {
		o.Level = l
		return nil
	}
}

// WithGlobPattern 日志滚动
func WithGlobPattern(g string) Option {
	return func(o *Options) error {
		o.GlobPattern = g
		return nil
	}
}

func WithLinkName(l string) Option {
	return func(o *Options) error {
		o.LinkName = l
		return nil
	}
}

func WithRotationSize(s int64) Option {
	return func(o *Options) error {
		o.RotationSize = s
		return nil
	}
}

func WithRotationTime(t time.Duration) Option {
	return func(o *Options) error {
		o.RotationTime = t
		return nil
	}
}

func WithMaxAge(a time.Duration) Option {
	return func(o *Options) error {
		o.LogMaxAge = a
		return nil
	}
}

func WithRotationCount(c uint) Option {
	return func(o *Options) error {
		o.RotationCount = c
		return nil
	}
}

func WithCaller(c bool) Option {
	return func(o *Options) error {
		o.Caller = c
		return nil
	}
}

func WithCallerFunc(c bool) Option {
	return func(o *Options) error {
		o.CallerFunc = c
		return nil
	}
}

func WithCallerSkip(c int) Option {
	return func(o *Options) error {
		o.CallerSkip = c
		return nil
	}
}

// options中GlobPattern不为空使用rotatelogs轮转
// 同时存在使用rotatelogs轮转
func NewServerLogger(tag string, opts ...Option) (*ServerLog, error) {
	l := &ServerLog{}
	logger, err := NewLogger(opts...)
	if err != nil {
		return nil, err
	}

	l.Helper = NewHelper(logger, tag)
	return l, nil
}

func NewLogger(opts ...Option) (*zap.Logger, error) {
	options, err := NewOptions(opts...)
	if err != nil {
		return nil, err
	}

	// 文件路径为空，只输出到控制台
	ws := zapcore.NewMultiWriteSyncer(zapcore.AddSync(os.Stdout))

	// rotatelogs
	if options.GlobPattern != "" {
		hook, err := NewRotation(options)
		if err != nil {
			return nil, err
		}
		if options.OutStd {
			ws = zapcore.NewMultiWriteSyncer(zapcore.AddSync(os.Stdout), zapcore.AddSync(hook))
		} else {
			ws = zapcore.NewMultiWriteSyncer(zapcore.AddSync(hook))
		}
	}

	core := zapcore.NewCore(
		zapcore.NewJSONEncoder(*options.EncoderConf), // 编码器配置
		ws,                                  // 输出方式
		zap.NewAtomicLevelAt(options.Level), // 日志级别
	)

	var logger *zap.Logger
	if options.Caller {
		if options.CallerFunc {
			options.EncoderConf.EncodeCaller = func(caller zapcore.EntryCaller, enc zapcore.PrimitiveArrayEncoder) {
				enc.AppendString(fmt.Sprintf("%s:%d %s()", path.Base(caller.File), caller.Line, funcShortName(caller.Function)))
			}
		} else {
			options.EncoderConf.EncodeCaller = func(caller zapcore.EntryCaller, enc zapcore.PrimitiveArrayEncoder) {
				enc.AppendString(fmt.Sprintf("%s:%d", path.Base(caller.File), caller.Line))
			}
		}
		options.EncoderConf.FunctionKey = zapcore.OmitKey
		logger = zap.New(core, zap.AddCallerSkip(options.CallerSkip), zap.AddCaller())
	} else {
		logger = zap.New(core)
	}
	return logger, err
}

func SetSLog(l *ServerLog) {
	sLog = l
}

func funcShortName(name string) string {
	for i := len(name) - 1; i >= 0 && name[i] != '/'; i-- {
		if name[i] == '.' {
			return name[i+1:]
		}
	}

	return name
}

func Sync() {
	sLog.Helper.Sync()
}
