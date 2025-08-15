package server

import (
	"context"
	"fmt"
	"github.com/pkg/errors"
	"github.com/sirupsen/logrus"
	"google.golang.org/grpc"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/status"
	"service1/internal/logctx"
	"time"
)

type ctxKeyLogger struct{}

func UnaryServerLogger(log *logrus.Logger) grpc.UnaryServerInterceptor {
	return func(
		ctx context.Context,
		req interface{},
		info *grpc.UnaryServerInfo,
		handler grpc.UnaryHandler,
	) (resp interface{}, err error) {

		// достаем/генерим request-id
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

		start := time.Now()
		entry.Info("rpc start")

		// положим entry в контекст, чтобы хэндлер мог достать
		ctx = context.WithValue(ctx, ctxKeyLogger{}, entry)

		resp, err = handler(ctx, req)

		st, _ := status.FromError(err)
		entry = entry.WithFields(logrus.Fields{
			"grpc_code": st.Code(),
			"duration":  time.Since(start).String(),
		})

		if err != nil {
			werr := errors.WithStack(err)
			entry.WithField("stack", fmt.Sprintf("%+v", werr)).WithError(err).Error("rpc end with error")
		} else {
			entry.Info("rpc end")
		}
		return resp, err
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
