package mw

import (
	"context"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

type ctxKey string

const CtxRequestID ctxKey = "x-request-id"
const HeaderRequestID = "X-Request-ID"

func EnsureRequestID(ctx context.Context, hdr string) (context.Context, string) {
	if hdr != "" {
		return context.WithValue(ctx, CtxRequestID, hdr), hdr
	}
	id := uuid.NewString()
	return context.WithValue(ctx, CtxRequestID, id), id
}

func FromContext(ctx context.Context) string {
	if v := ctx.Value(CtxRequestID); v != nil {
		if s, ok := v.(string); ok {
			return s
		}
	}
	return ""
}

func RequestID() gin.HandlerFunc {
	return func(c *gin.Context) {
		ctx, id := EnsureRequestID(c.Request.Context(), c.GetHeader(HeaderRequestID))
		c.Request = c.Request.WithContext(ctx)
		c.Writer.Header().Set(HeaderRequestID, id)
		c.Next()
	}
}
