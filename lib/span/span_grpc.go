package span

import (
	"context"
	"fmt"
	"strings"

	"github.com/opentracing/opentracing-go"
	"github.com/opentracing/opentracing-go/ext"
	"google.golang.org/grpc"
	"google.golang.org/grpc/grpclog"
	"google.golang.org/grpc/metadata"
)

// MDReaderWriter metadata 读写
type MDReaderWriter struct {
	metadata.MD
}

// ForeachKey 为了 opentracing.TextMapReader ，参考 opentracing 代码
func (c MDReaderWriter) ForeachKey(handler func(key, val string) error) error {
	for k, vs := range c.MD {
		for _, v := range vs {
			if err := handler(k, v); err != nil {
				return err
			}
		}
	}
	return nil
}

// Set 为了 opentracing.TextMapWriter，参考 opentracing 代码
func (c MDReaderWriter) Set(key, val string) {
	key = strings.ToLower(key)
	c.MD[key] = append(c.MD[key], val)
}

// ClientInterceptor rpc拦截器
func ClientInterceptor(tracer opentracing.Tracer) grpc.UnaryClientInterceptor {
	return func(ctx context.Context,
		method string,
		req,
		reply interface{},
		cc *grpc.ClientConn,
		invoker grpc.UnaryInvoker,
		opts ...grpc.CallOption) error {

		// 创建 rootSpan
		var rootCtx opentracing.SpanContext

		rootSpan := opentracing.SpanFromContext(ctx)
		if rootSpan != nil {
			rootCtx = rootSpan.Context()
		}

		span := tracer.StartSpan(
			method,
			opentracing.ChildOf(rootCtx),
			ext.SpanKindRPCClient,
		)

		defer span.Finish()

		md, succ := metadata.FromOutgoingContext(ctx)
		if !succ {
			md = metadata.New(nil)
		} else {
			md = md.Copy()
		}

		mdWriter := MDReaderWriter{md}

		// 注入 spanContext
		err := tracer.Inject(span.Context(), opentracing.TextMap, mdWriter)
		if err != nil {
			fmt.Println("tracer inject err", err.Error())
		}

		// new ctx ，并调用后续操作
		newCtx := metadata.NewOutgoingContext(ctx, md)
		return invoker(newCtx, method, req, reply, cc, opts...)
	}
}

// ServerInterceptor server端rpc拦截器
func ServerInterceptor(tracer opentracing.Tracer) grpc.UnaryServerInterceptor {
	return func(ctx context.Context,
		req interface{},
		info *grpc.UnaryServerInfo,
		handler grpc.UnaryHandler) (res interface{}, err error) {

		md, succ := metadata.FromIncomingContext(ctx)
		if !succ {
			md = metadata.New(nil)
		}

		// 提取 spanContext
		spanContext, err := tracer.Extract(opentracing.TextMap, MDReaderWriter{md})
		if err != nil && err != opentracing.ErrSpanContextNotFound {
			grpclog.Errorf("extract from metadata err: %v", err)
		} else {
			span := tracer.StartSpan(
				info.FullMethod,
				ext.RPCServerOption(spanContext),
				opentracing.Tag{Key: string(ext.Component), Value: "grpc"},
				ext.SpanKindRPCServer,
			)
			defer span.Finish()
			ctx = opentracing.ContextWithSpan(ctx, span)
		}
		return handler(ctx, req)
	}
}
