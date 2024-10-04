package log

import (
	"go.uber.org/zap"
)

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
