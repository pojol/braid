package grpc

import (
	"errors"
	"fmt"
	"net"

	grpc_middleware "github.com/grpc-ecosystem/go-grpc-middleware"
	"google.golang.org/grpc"
)

var (
	// ErrServiceUnavailiable 没有可用的服务
	ErrServiceUnavailiable = errors.New("service not registered")
)

// Server RPC 服务端
type Server struct {
	rpc *grpc.Server

	listen net.Listener
	parm   ServerParm
}

func BuildServerWithOption(opts ...ServerOption) *Server {

	p := ServerParm{
		ListenAddr: ":14222",
	}

	for _, opt := range opts {
		opt(&p)
	}

	var rpcserver *grpc.Server

	if len(p.UnaryInterceptors) != 0 {
		rpcserver = grpc.NewServer(grpc.UnaryInterceptor(grpc_middleware.ChainUnaryServer(p.UnaryInterceptors...)))
	} else {
		rpcserver = grpc.NewServer()
	}

	if p.Handler == nil {
		panic(fmt.Errorf("grpc server handler not set"))
	}

	return &Server{
		parm: p,
		rpc:  rpcserver,
	}

}

func (s *Server) Init() error {

	rpcListen, err := net.Listen("tcp", s.parm.ListenAddr)
	if err != nil {
		return fmt.Errorf("%v [GRPC] server check error %v [%v]", "", "tcp", s.parm.ListenAddr)
	} else {
		fmt.Printf("[GRPC] server listen: [tcp] %v\n", s.parm.ListenAddr)
	}

	s.listen = rpcListen

	return nil
}

// Get 获取rpc 服务器
func (s *Server) Server() interface{} {
	return s.rpc
}

// Run 运行
func (s *Server) Run() {

	// regist rpc handler
	s.parm.Handler(s.rpc)

	go func() {
		fmt.Println("[GRPC] server serving ...")
		if err := s.rpc.Serve(s.listen); err != nil {
			fmt.Printf("[GRPC] server serving err %s", err.Error())
		}
	}()

}

// Close 退出处理
func (s *Server) Close() {
	fmt.Printf("grpc-server closed")
	if s.parm.GracefulStop {
		s.rpc.GracefulStop()
	} else {
		s.rpc.Stop()
	}
}
