package grpc

import (
	"context"
	"errors"
	"fmt"
	"net"
	"testing"
	"time"

	"github.com/pojol/braid/lib/grpc/mock"
	"github.com/stretchr/testify/assert"
	"google.golang.org/grpc"
	"google.golang.org/grpc/test/bufconn"
)

// MockService
type MockService struct {
	delay     time.Duration
	shouldErr bool
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

func getBufDialer(lis *bufconn.Listener) func(context.Context, string) (net.Conn, error) {
	return func(ctx context.Context, url string) (net.Conn, error) {
		return lis.Dial()
	}
}

// create test gRPC server
func setupMockServer(t *testing.T, mockService *MockService) (*grpc.Server, *bufconn.Listener) {
	lis := bufconn.Listen(1024 * 1024)
	s := grpc.NewServer()
	// register mock service
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
					WithDialOptions(grpc.WithContextDialer(getBufDialer(lis))),
					WithClientConns([]string{"bufconn"}),
				)

				err := client.Init()
				if err != nil {
					t.Fatalf("Failed to initialize client: %v", err)
				}

				return client, srv, "bufconn"
			},
			ctx:    context.Background(),
			method: "/mock.MockService/Process",
			args:   &mock.MockRequest{Message: "test"},
		},
		{
			name: "context cancelled",
			setup: func() (*Client, *grpc.Server, string) {
				mockService := &MockService{}
				srv, lis := setupMockServer(t, mockService)
				client := BuildClientWithOption(
					WithMaxConcurrentCalls(1),
					WithCallTimeout(time.Second),
					WithDialOptions(grpc.WithContextDialer(getBufDialer(lis))),
					WithClientConns([]string{"bufconn"}),
				)

				// 初始化客户端
				err := client.Init()
				if err != nil {
					t.Fatalf("Failed to initialize client: %v", err)
				}

				return client, srv, "bufconn"
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
					WithDialOptions(grpc.WithContextDialer(getBufDialer(lis))),
					WithClientConns([]string{"bufconn"}),
				)

				// 初始化客户端
				err := client.Init()
				if err != nil {
					t.Fatalf("Failed to initialize client: %v", err)
				}

				return client, srv, "bufconn"
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

				// 初始化客户端
				err := client.Init()
				if err != nil {
					t.Fatalf("Failed to initialize client: %v", err)
				}
				return client, nil, "invalid:12345"
			},
			ctx:         context.Background(),
			method:      "/mock.MockService/Process",
			args:        &mock.MockRequest{Message: "test"},
			expectedErr: "gRPC client Can't find target invalid:12345",
		},
		{
			name: "concurrent calls limit",
			setup: func() (*Client, *grpc.Server, string) {
				mockService := &MockService{delay: time.Millisecond * 500}
				srv, lis := setupMockServer(t, mockService)
				client := BuildClientWithOption(
					WithMaxConcurrentCalls(1),
					WithCallTimeout(time.Second),
					WithDialOptions(grpc.WithContextDialer(getBufDialer(lis))),
					WithClientConns([]string{"bufconn"}),
				)

				// 初始化客户端
				err := client.Init()
				if err != nil {
					t.Fatalf("Failed to initialize client: %v", err)
				}

				return client, srv, "bufconn"
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

			err := client.Call(tt.ctx, addr, tt.method, tt.args, &mock.MockResponse{})

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
		WithMaxConcurrentCalls(500),
		WithCallTimeout(time.Second*3),
		WithDialOptions(grpc.WithContextDialer(getBufDialer(lis))),
		WithClientConns([]string{"bufconn"}),
	)

	err := client.Init()
	if err != nil {
		t.Fatalf("Failed to initialize client: %v", err)
	}

	concurrentCalls := 10000
	errChan := make(chan error, concurrentCalls)

	for i := 0; i < concurrentCalls; i++ {
		go func(i int) {
			ctx := context.Background()
			err := client.Call(ctx, "bufconn", "/mock.MockService/Process",
				&mock.MockRequest{Message: fmt.Sprintf("test-%d", i)}, &mock.MockResponse{})
			errChan <- err
		}(i)
	}

	var errors []error
	successCount := 0
	failureCount := 0

	for i := 0; i < concurrentCalls; i++ {
		if err := <-errChan; err != nil {
			errors = append(errors, err)
			failureCount++
		} else {
			successCount++
		}
	}

	t.Logf("Total calls: %d", concurrentCalls)
	t.Logf("Successful calls: %d", successCount)
	t.Logf("Failed calls: %d", failureCount)
	if len(errors) > 0 {
		t.Logf("Error samples: %v", errors[0])
	}

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
		WithDialOptions(grpc.WithContextDialer(getBufDialer(lis))),
		WithClientConns([]string{"bufconn"}),
	)

	err := client.Init()
	if err != nil {
		t.Fatalf("Failed to initialize client: %v", err)
	}

	err = client.Call(context.Background(), "bufconn",
		"/mock.MockService/Process", &mock.MockRequest{Message: "test"}, &mock.MockResponse{})
	assert.NoError(t, err)

	select {
	case client.workers <- struct{}{}:
		<-client.workers // 释放工作槽
	default:
		t.Error("Worker slot was not properly released")
	}
}
