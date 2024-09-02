package grpc

import "google.golang.org/grpc"

type RegistHandler func(*grpc.Server)

// Parm Service 配置
type ServerParm struct {
	ListenAddr string

	UnaryInterceptors  []grpc.UnaryServerInterceptor
	StreamInterceptors []grpc.StreamServerInterceptor

	Handler RegistHandler

	GracefulStop bool
}

// Option config wraps
type ServerOption func(*ServerParm)

// WithListen 服务器侦听地址配置
func WithServerListen(address string) ServerOption {
	return func(c *ServerParm) {
		c.ListenAddr = address
	}
}

func WithServerGracefulStop() ServerOption {
	return func(c *ServerParm) {
		c.GracefulStop = true
	}
}

func ServerAppendUnaryInterceptors(interceptor grpc.UnaryServerInterceptor) ServerOption {
	return func(c *ServerParm) {
		c.UnaryInterceptors = append(c.UnaryInterceptors, interceptor)
	}
}

func ServerAppendStreamInterceptors(interceptor grpc.StreamServerInterceptor) ServerOption {
	return func(c *ServerParm) {
		c.StreamInterceptors = append(c.StreamInterceptors, interceptor)
	}
}

func ServerRegisterHandler(handler RegistHandler) ServerOption {
	return func(c *ServerParm) {
		c.Handler = handler
	}
}
