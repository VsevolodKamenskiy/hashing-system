package grpcclient

import (
	"context"

	"google.golang.org/grpc"
	"google.golang.org/grpc/metadata"

	"service2/internal/mw"
)

func UnaryClientInjectRequestID() grpc.UnaryClientInterceptor {
	return func(
		ctx context.Context,
		method string,
		req, reply interface{},
		cc *grpc.ClientConn,
		invoker grpc.UnaryInvoker,
		opts ...grpc.CallOption,
	) error {
		id := mw.FromContext(ctx)
		md, _ := metadata.FromOutgoingContext(ctx)
		md = md.Copy()
		if id != "" {
			md.Set("x-request-id", id)
		}
		ctx = metadata.NewOutgoingContext(ctx, md)
		return invoker(ctx, method, req, reply, cc, opts...)
	}
}
