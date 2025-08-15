package server

import (
	"context"
	"time"

	"github.com/prometheus/client_golang/prometheus"
	"google.golang.org/grpc"
	"google.golang.org/grpc/status"
)

var (
	rpcRequestsTotal = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Name: "service1_grpc_requests_total",
			Help: "Total number of gRPC requests handled by service1.",
		},
		[]string{"method", "code"},
	)
	rpcRequestDuration = prometheus.NewHistogramVec(
		prometheus.HistogramOpts{
			Name:    "service1_grpc_request_duration_seconds",
			Help:    "Duration of gRPC requests handled by service1.",
			Buckets: prometheus.DefBuckets,
		},
		[]string{"method", "code"},
	)
)

func init() {
	prometheus.MustRegister(rpcRequestsTotal, rpcRequestDuration)
}

func UnaryServerMetrics() grpc.UnaryServerInterceptor {
	return func(ctx context.Context, req interface{}, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (interface{}, error) {
		start := time.Now()
		resp, err := handler(ctx, req)
		code := status.Code(err).String()
		rpcRequestsTotal.WithLabelValues(info.FullMethod, code).Inc()
		rpcRequestDuration.WithLabelValues(info.FullMethod, code).Observe(time.Since(start).Seconds())
		return resp, err
	}
}
