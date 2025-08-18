package server

import (
	prom "github.com/grpc-ecosystem/go-grpc-middleware/providers/prometheus"
)

// NewServerMetrics creates Prometheus metrics collector for gRPC server.
func NewServerMetrics() *prom.ServerMetrics {
	return prom.NewServerMetrics(
		prom.WithServerHandlingTimeHistogram(),
	)
}
