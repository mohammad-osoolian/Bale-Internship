package server

import (
	"context"
	"therealbroker/api/metrics"
	"time"

	"google.golang.org/grpc"
	"google.golang.org/grpc/status"
)

// UnaryServerInterceptor for measuring the duration of unary RPC calls.
func UnaryMetricsInterceptor() grpc.UnaryServerInterceptor {
	return func(
		ctx context.Context,
		req interface{},
		info *grpc.UnaryServerInfo,
		handler grpc.UnaryHandler,
	) (interface{}, error) {
		start := time.Now()
		resp, err := handler(ctx, req)

		duration := time.Since(start).Seconds()
		metrics.RpcDurations.WithLabelValues(info.FullMethod).Observe(duration)

		code := status.Code(err)
		metrics.RpcCalls.WithLabelValues(code.String(), info.FullMethod).Inc()

		return resp, err
	}
}

// StreamServerInterceptor for measuring the duration of streaming RPC calls.
func StreamMetricsInterceptor() grpc.StreamServerInterceptor {
	return func(
		srv interface{},
		stream grpc.ServerStream,
		info *grpc.StreamServerInfo,
		handler grpc.StreamHandler,
	) error {
		metrics.ActiveSubscriptions.Inc()
		defer metrics.ActiveSubscriptions.Dec()

		start := time.Now()
		err := handler(srv, stream)

		duration := time.Since(start).Seconds()
		metrics.RpcDurations.WithLabelValues(info.FullMethod).Observe(duration)

		code := status.Code(err)
		metrics.RpcCalls.WithLabelValues(code.String(), info.FullMethod).Inc()

		return err
	}
}
