package server

import (
	"context"

	"github.com/grpc-ecosystem/go-grpc-middleware/v2/interceptors/logging"
	"github.com/sirupsen/logrus"
	"google.golang.org/grpc"
	"google.golang.org/grpc/metadata"

	"service1/internal/logctx"
)

type ctxKeyLogger struct{}

// UnaryRequestID ensures that every request has a request id and injects it
// into both context and logging fields. It also stores a logrus entry with the
// request id in the context so that handlers can retrieve it.
func UnaryRequestID(log *logrus.Logger) grpc.UnaryServerInterceptor {
	return func(ctx context.Context, req interface{}, _ *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (interface{}, error) {
		if md, ok := metadata.FromIncomingContext(ctx); ok {
			if vals := md.Get(logctx.RequestIDKey); len(vals) > 0 && vals[0] != "" {
				ctx = context.WithValue(ctx, logctx.RequestIDKey, vals[0])
			}
		}

		ctx, reqID := logctx.EnsureRequestID(ctx)
		entry := log.WithFields(logrus.Fields{
			"request_id": reqID,
			"component":  "service1",
		})

		// store entry for handlers and inject fields for logging interceptor
		ctx = context.WithValue(ctx, ctxKeyLogger{}, entry)
		ctx = logging.InjectFields(ctx, logging.Fields{
			"request_id", reqID,
			"component", "service1",
		})

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

// LoggingInterceptor returns a unary server interceptor from the
// go-grpc-middleware logging package configured with the provided logger.
func LoggingInterceptor(log *logrus.Logger) grpc.UnaryServerInterceptor {
	logger := logging.LoggerFunc(func(ctx context.Context, level logging.Level, msg string, fields ...any) {
		entry := log.WithContext(ctx)
		for i := 0; i+1 < len(fields); i += 2 {
			key, ok := fields[i].(string)
			if !ok {
				continue
			}
			entry = entry.WithField(key, fields[i+1])
		}
		switch level {
		case logging.LevelDebug:
			entry.Debug(msg)
		case logging.LevelInfo:
			entry.Info(msg)
		case logging.LevelWarn:
			entry.Warn(msg)
		case logging.LevelError:
			entry.Error(msg)
		default:
			entry.Info(msg)
		}
	})
	return logging.UnaryServerInterceptor(logger,
		logging.WithFieldsFromContext(logging.ExtractFields),
		logging.WithLogOnEvents(logging.StartCall, logging.FinishCall),
	)
}
