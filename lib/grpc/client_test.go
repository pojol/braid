package grpc

import (
	"context"
	"errors"
	"fmt"
	"testing"
	"time"

	"github.com/pojol/braid/lib/grpc/mock"
	"github.com/stretchr/testify/assert"
	"google.golang.org/grpc"
	"google.golang.org/grpc/test/bufconn"
)

// MockService 用于测试的模拟服务
type MockService struct {
	delay     time.Duration // 用于模拟处理延迟
	shouldErr bool          // 是否应该返回错误
	mock.MockServiceServer
}

func (s *MockService) Process(ctx context.Context, req *mock.MockRequest) (*mock.MockResponse, error) {
	if s.delay > 0 {
		time.Sleep(s.delay)
	}
	if s.shouldErr {
		return nil, errors.New("mock error")
	}
	return &mock.MockResponse{Message: "ok"}, nil
}

// 创建测试用的 gRPC server
func setupMockServer(t *testing.T, mockService *MockService) (*grpc.Server, *bufconn.Listener) {
	lis := bufconn.Listen(1024 * 1024)
	s := grpc.NewServer()
	// 注册 mock service
	mock.RegisterMockServiceServer(s, mockService)
	go func() {
		if err := s.Serve(lis); err != nil {
			t.Errorf("failed to serve: %v", err)
		}
	}()
	return s, lis
}

func TestClient_Call(t *testing.T) {
	tests := []struct {
		name        string
		setup       func() (*Client, *grpc.Server, string)
		ctx         context.Context
		method      string
		args        interface{}
		opts        []interface{}
		delay       time.Duration
		shouldErr   bool
		expectedErr string
	}{
		{
			name: "successful call",
			setup: func() (*Client, *grpc.Server, string) {
				mockService := &MockService{}
				srv, lis := setupMockServer(t, mockService)
				client := BuildClientWithOption(
					WithMaxConcurrentCalls(1),
					WithCallTimeout(time.Second),
				)
				return client, srv, lis.Addr().String()
			},
			ctx:    context.Background(),
			method: "/mock.MockService/Process",
			args:   &mock.MockRequest{Message: "test"},
		},
		{
			name: "timeout",
			setup: func() (*Client, *grpc.Server, string) {
				mockService := &MockService{delay: time.Second * 2}
				srv, lis := setupMockServer(t, mockService)
				client := BuildClientWithOption(
					WithMaxConcurrentCalls(1),
					WithCallTimeout(time.Second),
				)
				return client, srv, lis.Addr().String()
			},
			ctx:         context.Background(),
			method:      "/mock.MockService/Process",
			args:        &mock.MockRequest{Message: "test"},
			expectedErr: "call timeout after 1s",
		},
		{
			name: "context cancelled",
			setup: func() (*Client, *grpc.Server, string) {
				mockService := &MockService{}
				srv, lis := setupMockServer(t, mockService)
				client := BuildClientWithOption(
					WithMaxConcurrentCalls(1),
					WithCallTimeout(time.Second),
				)
				return client, srv, lis.Addr().String()
			},
			ctx: func() context.Context {
				ctx, cancel := context.WithCancel(context.Background())
				cancel()
				return ctx
			}(),
			method:      "/mock.MockService/Process",
			args:        &mock.MockRequest{Message: "test"},
			expectedErr: "context canceled",
		},
		{
			name: "service error",
			setup: func() (*Client, *grpc.Server, string) {
				mockService := &MockService{shouldErr: true}
				srv, lis := setupMockServer(t, mockService)
				client := BuildClientWithOption(
					WithMaxConcurrentCalls(1),
					WithCallTimeout(time.Second),
				)
				return client, srv, lis.Addr().String()
			},
			ctx:         context.Background(),
			method:      "/mock.MockService/Process",
			args:        &mock.MockRequest{Message: "test"},
			expectedErr: "mock error",
		},
		{
			name: "invalid connection",
			setup: func() (*Client, *grpc.Server, string) {
				client := BuildClientWithOption(
					WithMaxConcurrentCalls(1),
					WithCallTimeout(time.Second),
				)
				return client, nil, "invalid:12345"
			},
			ctx:         context.Background(),
			method:      "/mock.MockService/Process",
			args:        &mock.MockRequest{Message: "test"},
			expectedErr: "connection error",
		},
		{
			name: "concurrent calls limit",
			setup: func() (*Client, *grpc.Server, string) {
				mockService := &MockService{delay: time.Millisecond * 500}
				srv, lis := setupMockServer(t, mockService)
				client := BuildClientWithOption(
					WithMaxConcurrentCalls(1),
					WithCallTimeout(time.Second),
				)
				return client, srv, lis.Addr().String()
			},
			ctx: func() context.Context {
				ctx, cancel := context.WithTimeout(context.Background(), time.Second*2)
				go func() {
					time.Sleep(time.Second * 3)
					cancel()
				}()
				return ctx
			}(),
			method: "/mock.MockService/Process",
			args:   &mock.MockRequest{Message: "test"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			client, srv, addr := tt.setup()
			if srv != nil {
				defer srv.Stop()
			}

			err := client.Call(tt.ctx, addr, tt.method, tt.args, tt.opts...)

			if tt.expectedErr != "" {
				assert.Error(t, err)
				assert.Contains(t, err.Error(), tt.expectedErr)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

// 测试并发调用
func TestClient_Call_Concurrent(t *testing.T) {
	mockService := &MockService{delay: time.Millisecond * 100}
	srv, lis := setupMockServer(t, mockService)
	defer srv.Stop()

	client := BuildClientWithOption(
		WithMaxConcurrentCalls(5),
		WithCallTimeout(time.Second),
	)

	concurrentCalls := 10
	errChan := make(chan error, concurrentCalls)

	// 并发发起调用
	for i := 0; i < concurrentCalls; i++ {
		go func(i int) {
			ctx := context.Background()
			err := client.Call(ctx, lis.Addr().String(), "/mock.MockService/Process",
				&mock.MockRequest{Message: fmt.Sprintf("test-%d", i)})
			errChan <- err
		}(i)
	}

	// 收集结果
	var errors []error
	for i := 0; i < concurrentCalls; i++ {
		if err := <-errChan; err != nil {
			errors = append(errors, err)
		}
	}

	// 验证结果
	assert.Less(t, len(errors), concurrentCalls/2,
		"Too many errors in concurrent calls")
}

// 测试资源清理
func TestClient_Call_ResourceCleanup(t *testing.T) {
	mockService := &MockService{}
	srv, lis := setupMockServer(t, mockService)
	defer srv.Stop()

	client := BuildClientWithOption(
		WithMaxConcurrentCalls(1),
		WithCallTimeout(time.Second),
	)

	// 执行调用
	err := client.Call(context.Background(), lis.Addr().String(),
		"/mock.MockService/Process", &mock.MockRequest{Message: "test"})
	assert.NoError(t, err)

	// 验证工作槽是否被正确释放
	select {
	case client.workers <- struct{}{}:
		// 如果可以获取工作槽，说明之前的调用正确释放了资源
		<-client.workers // 释放工作槽
	default:
		t.Error("Worker slot was not properly released")
	}
}
