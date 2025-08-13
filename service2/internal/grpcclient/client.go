package grpcclient

import (
	"context"

	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"

	hasherpb "service2/proto/hasherpb"
)

type HasherClient interface {
	Calculate(ctx context.Context, strings []string) ([]string, error)
}

type client struct {
	conn *grpc.ClientConn
	c    hasherpb.HasherServiceClient
}

func New(addr string, extra ...grpc.DialOption) (HasherClient, error) {
	opts := []grpc.DialOption{
		grpc.WithTransportCredentials(insecure.NewCredentials()),
		grpc.WithChainUnaryInterceptor(UnaryClientInjectRequestID()),
	}
	opts = append(opts, extra...)

	conn, err := grpc.Dial(addr, opts...)
	if err != nil {
		return nil, err
	}
	return &client{conn: conn, c: hasherpb.NewHasherServiceClient(conn)}, nil
}

func (cl *client) Calculate(ctx context.Context, strings []string) ([]string, error) {
	resp, err := cl.c.CalculateHashes(ctx, &hasherpb.HashRequest{Strings: strings})
	if err != nil {
		return nil, err
	}
	return resp.GetHashes(), nil
}
