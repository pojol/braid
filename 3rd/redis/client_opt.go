package redis

import (
	"time"

	"github.com/pojol/braid/lib/tracer"
	"github.com/redis/go-redis/v9"
)

// Parm 配置项
type Parm struct {
	redisopt redis.Options
	trc      tracer.ITracer
}

// Option config wraps
type Option func(options *Parm)

// WithAddr redis 连接地址 "redis:// :password@10.0.1.11:6379/0"
func WithAddr(addr string) Option {
	return func(c *Parm) {
		p, err := redis.ParseURL(addr)
		if err != nil {
			panic(err)
		}
		c.redisopt = *p
	}
}

// 通过域名进行访问
func WithDomain(domain string) Option {
	return func(c *Parm) {
		c.redisopt.Addr = domain
	}
}

func WithUsername(name string) Option {
	return func(c *Parm) {
		c.redisopt.Username = name
	}
}

func WithPassword(pwd string) Option {
	return func(c *Parm) {
		c.redisopt.Password = pwd
	}
}

// WithReadTimeOut 连接的读取超时时间
func WithReadTimeOut(readtimeout time.Duration) Option {
	return func(c *Parm) {
		c.redisopt.ReadTimeout = readtimeout
	}
}

// WithWriteTimeOut 连接的写入超时时间
func WithWriteTimeOut(writetimeout time.Duration) Option {
	return func(c *Parm) {
		c.redisopt.WriteTimeout = writetimeout
	}
}

// WithConnectTimeOut 连接超时时间
func WithConnectTimeOut(connecttimeout time.Duration) Option {
	return func(c *Parm) {
		c.redisopt.ConnMaxLifetime = connecttimeout
	}
}

// WithIdleTimeout 闲置连接的超时时间, 设置小于服务器的超时时间 redis.conf : timeout
func WithIdleTimeout(idletimeout time.Duration) Option {
	return func(c *Parm) {
		c.redisopt.ConnMaxIdleTime = idletimeout
	}
}

// WithMaxIdle 最大空闲连接数
func WithMaxIdle(maxidle int) Option {
	return func(c *Parm) {
		c.redisopt.MaxIdleConns = maxidle
	}
}

func WithMinIdle(minidle int) Option {
	return func(options *Parm) {
		options.redisopt.MinIdleConns = minidle
	}
}

// WithPoolSize 链接数
func WithPoolSize(size int) Option {
	return func(c *Parm) {
		c.redisopt.PoolSize = size
	}
}

func WithPoolTimeout(timeout time.Duration) Option {
	return func(options *Parm) {
		options.redisopt.PoolTimeout = timeout
	}
}

func WithTracer(trc tracer.ITracer) Option {
	return func(c *Parm) {
		c.trc = trc
	}
}
