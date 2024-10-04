package grpc

import (
	"context"
	"errors"
	"fmt"
	"sync"
	"time"

	grpc_middleware "github.com/grpc-ecosystem/go-grpc-middleware"

	"google.golang.org/grpc"
	"google.golang.org/grpc/connectivity"
)

var (
	// ErrServiceNotAvailable 服务不可用，通常是因为没有查询到中心节点(coordinate)
	ErrServiceNotAvailable = errors.New("caller service not available")

	// ErrConfigConvert 配置转换失败
	ErrConfigConvert = errors.New("convert linker config")

	// ErrCantFindNode 在注册中心找不到对应的服务节点
	ErrCantFindNode = errors.New("can't find service node in center")
)

// Client 调用器
type Client struct {
	parm    ClientParm
	connmap sync.Map
}

func BuildClientWithOption(opts ...ClientOption) *Client {

	p := DefaultClientParm

	for _, opt := range opts {
		opt(&p)
	}

	return &Client{
		parm: p,
	}
}

func (c *Client) newconn(addr string) (*grpc.ClientConn, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	var conn *grpc.ClientConn
	var err error

	if len(c.parm.UnaryInterceptors) > 0 {
		conn, err = grpc.DialContext(ctx, addr, grpc.WithInsecure(), grpc.WithUnaryInterceptor(grpc_middleware.ChainUnaryClient(c.parm.UnaryInterceptors...)))
		if err != nil {
			goto EXT
		}
	} else {
		conn, err = grpc.DialContext(ctx, addr, grpc.WithInsecure())
		if err != nil {
			goto EXT
		}
	}

EXT:
	//c.log.InfoFf("[braid.client] new connect addr : %v err : %v", addr, err)
	fmt.Printf("[braid.client] new connect addr : %v err : %v\n", addr, err)

	return conn, err
}

func (c *Client) closeconn(conn *grpc.ClientConn) error {
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	doneCh := make(chan error)
	go func() {
		var result error
		if err := conn.Close(); err != nil {
			result = fmt.Errorf("[braid.client] %w %v", err, "failed to close gRPC client")
		}
		doneCh <- result
	}()

	select {
	case <-ctx.Done():
		return errors.New("failed to close gRPC client because of timeout")
	case err := <-doneCh:
		fmt.Printf("[braid.client] close connect addr : %v err : %v", conn.Target(), err)
		return err
	}
}

func (c *Client) Init() error {

	for _, addr := range c.parm.AddressLst {
		conn, err := c.newconn(addr)
		if err != nil {
			fmt.Printf("[braid.client] new grpc conn err %s", err.Error())
		} else {
			c.connmap.Store(addr, conn)
		}
	}

	return nil
}

func (c *Client) getConn(address string) (*grpc.ClientConn, error) {
	mc, ok := c.connmap.Load(address)
	if !ok {
		return nil, fmt.Errorf("gRPC client Can't find target %s", address)
	}

	conn, ok := mc.(*grpc.ClientConn)
	if !ok {
		return nil, fmt.Errorf("gRPC client failed address : %s", address)
	}

	if conn.GetState() == connectivity.TransientFailure {
		fmt.Printf("[braid.client] reset connect backoff")
		conn.ResetConnectBackoff()
	}

	return conn, nil
}

func (c *Client) CallWait(ctx context.Context, addr, methon string, args, reply interface{}, opts ...interface{}) error {

	var grpcopts []grpc.CallOption

	conn, err := c.getConn(addr)
	if err != nil {
		// try create
		conn, err = c.newconn(addr)
		if err != nil {
			fmt.Printf("[braid.client] client get conn warning %s", err.Error())
			return err
		}

		c.connmap.Store(addr, conn)
	}

	if len(opts) != 0 {
		for _, v := range opts {
			callopt, ok := v.(grpc.CallOption)
			if !ok {
				fmt.Printf("[braid.client] call option type mismatch")
			}
			grpcopts = append(grpcopts, callopt)
		}
	}

	err = conn.Invoke(ctx, methon, args, reply, grpcopts...)
	if err != nil {
		fmt.Printf("[braid.client] invoke warning %s, methon = %s, addr = %s\n", err.Error(), methon, addr)
	}

	return err
}

func (c *Client) Call(ctx context.Context, addr, methon string, args interface{}, opts ...interface{}) error {

	var grpcopts []grpc.CallOption

	conn, err := c.getConn(addr)
	if err != nil {
		fmt.Printf("[braid.client] client get conn warning %s", err.Error())
		return err
	}

	if len(opts) != 0 {
		for _, v := range opts {
			callopt, ok := v.(grpc.CallOption)
			if !ok {
				fmt.Printf("[braid.client] call option type mismatch")
			}
			grpcopts = append(grpcopts, callopt)
		}
	}

	go func() {
		err = conn.Invoke(ctx, methon, args, nil, grpcopts...)
		if err != nil {
			fmt.Printf("[braid.client] invoke warning %s, methon = %s, addr = %s", err.Error(), methon, addr)
		}
	}()

	return err
}
