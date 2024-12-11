package dismutex

import (
	"context"
	"crypto/rand"
	"encoding/base64"
	"errors"
	"time"

	trdredis "github.com/pojol/braid/3rd/redis"
	"github.com/redis/go-redis/v9"
)

const (
	// Expiry 默认2秒的超时时间，当到达超时时强行释放锁。
	Expiry = 10 * time.Second

	// Tries 如果获取锁失败，可重试的次数
	Tries = 4

	// Delay 重新获得锁的间隔(毫秒
	Delay = 700
)

var (
	// ErrFailed is returned when lock cannot be acquired
	ErrFailed = errors.New("failed to acquire lock")
)

// Lock 分布式锁
func Lock(ctx context.Context, token string) (string, error) {

	if token == "" {
		return "", errors.New("empty token")
	}

	b := make([]byte, 16)
	_, err := rand.Read(b)
	if err != nil {
		return "", err
	}

	value := base64.StdEncoding.EncodeToString(b)
	for i := 0; i < Tries; i++ {
		reply, err := trdredis.SetNx(ctx, token, value, Expiry).Result()
		if err == nil && reply {
			return value, nil
		}

		time.Sleep(time.Duration(time.Millisecond * Delay))
	}

	return "", ErrFailed
}

// Unlock 释放锁
func Unlock(ctx context.Context, token, value string) bool {

	if value == "" {
		return false
	}

	status, err := delScript.Run(ctx, trdredis.GetClient(), []string{token}, value).Int()

	return status != 0 && err == nil
}

var delScript = redis.NewScript(`
if redis.call("get", KEYS[1]) == ARGV[1] then
	return redis.call("del", KEYS[1])
else
	return 0
end`)
