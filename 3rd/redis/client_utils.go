package redis

import (
	"context"
	"fmt"

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
	return nil, fmt.Errorf("nil")
}
