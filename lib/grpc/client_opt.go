package grpc

import (
	"time"

	"google.golang.org/grpc"
)

// Parm 调用器配置项
type ClientParm struct {
	PoolInitNum  int
	PoolCapacity int
	PoolIdle     time.Duration

	MaxConcurrentCalls int
	CallTimeout        time.Duration

	AddressLst []string

	dialOptions []grpc.DialOption

	UnaryInterceptors  []grpc.UnaryClientInterceptor
	StreamInterceptors []grpc.StreamClientInterceptor
}

var (
	DefaultClientParm = ClientParm{
		PoolInitNum:        8,
		PoolCapacity:       64,
		MaxConcurrentCalls: 1024,
		CallTimeout:        time.Second * 10,
		PoolIdle:           time.Second * 100,
	}
)

// Option config wraps
type ClientOption func(*ClientParm)

// WithPoolInitNum 连接池初始化数量
func WithClientPoolInitNum(num int) ClientOption {
	return func(c *ClientParm) {
		c.PoolInitNum = num
	}
}

// WithPoolCapacity 连接池的容量大小
func WithClientPoolCapacity(num int) ClientOption {
	return func(c *ClientParm) {
		c.PoolCapacity = num
	}
}

// WithPoolIdle 连接池的最大闲置时间
func WithClientPoolIdle(second int) ClientOption {
	return func(c *ClientParm) {
		c.PoolIdle = time.Duration(second) * time.Second
	}
}

// WithClientConns 目标服务器列表地址（静态绑定
func WithClientConns(lst []string) ClientOption {
	return func(c *ClientParm) {
		c.AddressLst = append(c.AddressLst, lst...)
	}
}

func WithMaxConcurrentCalls(maxCalls int) ClientOption {
	return func(cp *ClientParm) {
		cp.MaxConcurrentCalls = maxCalls
	}
}

func WithCallTimeout(callTimeout time.Duration) ClientOption {
	return func(cp *ClientParm) {
		cp.CallTimeout = callTimeout
	}
}

func WithDialOptions(opts ...grpc.DialOption) ClientOption {
	return func(cp *ClientParm) {
		cp.dialOptions = append(cp.dialOptions, opts...)
	}
}

func ClientAppendUnaryInterceptors(interceptor grpc.UnaryClientInterceptor) ClientOption {
	return func(c *ClientParm) {
		c.UnaryInterceptors = append(c.UnaryInterceptors, interceptor)
	}
}

func ClientAppendStreamInterceptors(interceptor grpc.StreamClientInterceptor) ClientOption {
	return func(c *ClientParm) {
		c.StreamInterceptors = append(c.StreamInterceptors, interceptor)
	}
}
