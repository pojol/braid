package redis

import (
	"context"
	"fmt"

	"github.com/pojol/braid/lib/span"
	"github.com/pojol/braid/lib/tracer"
	"github.com/redis/go-redis/v9"
)

func GetCmdsByteSlice(cmds []redis.Cmder) ([][]byte, error) {
	var bts [][]byte
	var err error
	for _, cmder := range cmds {
		cmd := cmder.(*redis.StringCmd)
		bytes, err := cmd.Bytes()
		if err != nil {
			goto EXT
		}
		bts = append(bts, bytes)
	}
EXT:
	return bts, err
}

type spanTag struct {
	key   string
	value string
}

func doTracing(ctx context.Context, args ...spanTag) (ispan tracer.ISpan, err error) {
	err = fmt.Errorf("tracer not init")
	if defaultClientConfig.trc != nil {
		ispan, err = defaultClientConfig.trc.GetSpan(span.RedisSpan)
		if err == nil {
			ispan.Begin(ctx)
			for _, v := range args {
				ispan.SetTag(v.key, v.value)
			}
		}
	}
	return
}
