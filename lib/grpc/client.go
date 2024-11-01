package grpc

import (
	"context"
	"errors"
	"fmt"
	"sync"
	"time"

	grpc_middleware "github.com/grpc-ecosystem/go-grpc-middleware"
	"github.com/pojol/braid/lib/log"
	"golang.org/x/sync/errgroup"

	"google.golang.org/grpc"
	"google.golang.org/grpc/connectivity"
	"google.golang.org/grpc/credentials/insecure"
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
	workers chan struct{} // 用于限制并发的 channel
}

func BuildClientWithOption(opts ...ClientOption) *Client {

	p := DefaultClientParm

	for _, opt := range opts {
		opt(&p)
	}

	return &Client{
		parm:    p,
		workers: make(chan struct{}, p.MaxConcurrentCalls), // 设置最大并发数
	}
}

func (c *Client) newconn(addr string) (*grpc.ClientConn, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	var conn *grpc.ClientConn
	var err error

	dialOpts := []grpc.DialOption{grpc.WithTransportCredentials(insecure.NewCredentials())}
	if len(c.parm.UnaryInterceptors) > 0 {
		dialOpts = append(dialOpts, grpc.WithUnaryInterceptor(grpc_middleware.ChainUnaryClient(c.parm.UnaryInterceptors...)))
	}
	if len(c.parm.dialOptions) > 0 {
		dialOpts = append(dialOpts, c.parm.dialOptions...)
	}

	conn, err = grpc.DialContext(ctx, addr, dialOpts...)
	if err != nil {
		log.WarnF("[braid.client] new connect addr : %v err : %v", addr, err)
		return nil, err
	}

	return conn, nil
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

func (c *Client) Call(ctx context.Context, addr, methon string, args interface{}, reply interface{}, opts ...interface{}) error {
	select {
	case c.workers <- struct{}{}: // 获取工作槽
		defer func() { <-c.workers }() // 释放工作槽
	case <-ctx.Done():
		return ctx.Err()
	}

	var grpcopts []grpc.CallOption

	conn, err := c.getConn(addr)
	if err != nil {
		log.WarnF("[braid.client] client get conn warning %s", err.Error())
		return err
	}

	if len(opts) != 0 {
		for _, v := range opts {
			callopt, ok := v.(grpc.CallOption)
			if !ok {
				log.WarnF("[braid.client] call option type mismatch")
			}
			grpcopts = append(grpcopts, callopt)
		}
	}

	// 使用 errgroup 来管理 goroutine
	g, ctx := errgroup.WithContext(ctx)
	g.Go(func() error {
		if err := conn.Invoke(ctx, methon, args, reply, grpcopts...); err != nil {
			return fmt.Errorf("[braid.client] invoke error: method=%s, addr=%s: %w",
				methon, addr, err)
		}
		return nil
	})

	// 设置超时
	done := make(chan error, 1)
	go func() {
		done <- g.Wait()
	}()

	select {
	case err := <-done:
		return err
	case <-time.After(c.parm.CallTimeout):
		return fmt.Errorf("call timeout after %v", c.parm.CallTimeout)
	}
}
