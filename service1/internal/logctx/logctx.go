package logctx

import (
	"context"
	"github.com/google/uuid"
	"google.golang.org/grpc/metadata"
)

const RequestIDKey = "x-request-id"

// EnsureRequestID возвращает существующий reqID из ctx или создает новый.
func EnsureRequestID(ctx context.Context) (context.Context, string) {
	// gRPC metadata -> x-request-id
	if md, ok := metadata.FromIncomingContext(ctx); ok {
		if vals := md.Get(RequestIDKey); len(vals) > 0 && vals[0] != "" {
			return ctx, vals[0]
		}
	}
	// обычный контекст (например, пришел из HTTP-мидлвари в service2)
	if v := ctx.Value(RequestIDKey); v != nil {
		if s, ok := v.(string); ok && s != "" {
			return ctx, s
		}
	}
	id := uuid.NewString()
	ctx = context.WithValue(ctx, RequestIDKey, id)
	return ctx, id
}
