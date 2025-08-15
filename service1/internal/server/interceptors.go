package server

import (
	"context"

	"github.com/sirupsen/logrus"
	"google.golang.org/grpc"
	"google.golang.org/grpc/metadata"
	"service1/internal/logctx"
)

type ctxKeyLogger struct{}

func UnaryServerRequestID(log *logrus.Logger) grpc.UnaryServerInterceptor {
	return func(
		ctx context.Context,
		req interface{},
		info *grpc.UnaryServerInfo,
		handler grpc.UnaryHandler,
	) (interface{}, error) {
		if md, ok := metadata.FromIncomingContext(ctx); ok {
			if vals := md.Get(logctx.RequestIDKey); len(vals) > 0 && vals[0] != "" {
				ctx = context.WithValue(ctx, logctx.RequestIDKey, vals[0])
			}
		}
		ctx, reqID := logctx.EnsureRequestID(ctx)

		entry := log.WithFields(logrus.Fields{
			"request_id": reqID,
			"rpc":        info.FullMethod,
			"component":  "service1",
		})

		ctx = context.WithValue(ctx, ctxKeyLogger{}, entry)

		return handler(ctx, req)
	}
}

func GetLoggerFromCtx(ctx context.Context, base *logrus.Logger) *logrus.Entry {
	if v := ctx.Value(ctxKeyLogger{}); v != nil {
		if e, ok := v.(*logrus.Entry); ok {
			return e
		}
	}
	_, reqID := logctx.EnsureRequestID(ctx)
	return base.WithField("request_id", reqID)
}
