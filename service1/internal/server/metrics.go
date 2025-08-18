package server

import (
	prom "github.com/grpc-ecosystem/go-grpc-middleware/providers/prometheus"
)

func NewServerMetrics() *prom.ServerMetrics {
	return prom.NewServerMetrics(
		prom.WithServerHandlingTimeHistogram(),
	)
}
