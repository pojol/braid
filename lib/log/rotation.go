package log

import (
	"io"

	rotatelogs "github.com/lestrrat-go/file-rotatelogs"
)

func NewRotation(options *Options) (io.Writer, error) {
	return rotatelogs.New(
		options.GlobPattern,
		rotatelogs.WithLinkName(options.LinkName),
		rotatelogs.WithRotationSize(options.RotationSize),
		rotatelogs.WithRotationTime(options.RotationTime),
		rotatelogs.WithMaxAge(options.LogMaxAge),
		rotatelogs.WithRotationCount(options.RotationCount),
	)
}
