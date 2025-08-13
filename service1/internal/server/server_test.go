package server_test

import (
	"context"
	"github.com/sirupsen/logrus"
	"net"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
	"google.golang.org/grpc"
	"google.golang.org/grpc/test/bufconn"

	"service1/internal/server"
	"service1/proto/hasherpb"
)

const bufSize = 1024 * 1024

func startBufGRPC(t *testing.T) (*grpc.ClientConn, func()) {
	t.Helper()

	logger := logrus.New()

	lis := bufconn.Listen(bufSize)
	s := grpc.NewServer()
	srv := &server.Server{Log: logger}
	hasherpb.RegisterHasherServiceServer(s, srv)

	go func() { _ = s.Serve(lis) }()

	dialer := func(context.Context, string) (net.Conn, error) { return lis.Dial() }
	conn, err := grpc.DialContext(context.Background(), "bufnet",
		grpc.WithContextDialer(dialer),
		grpc.WithInsecure(),
	)
	require.NoError(t, err)

	cleanup := func() {
		s.GracefulStop()
		_ = lis.Close()
		_ = conn.Close()
	}
	return conn, cleanup
}

func TestCalculateHashes_Basic(t *testing.T) {
	conn, cleanup := startBufGRPC(t)
	defer cleanup()

	client := hasherpb.NewHasherServiceClient(conn)

	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()

	in := []string{"a", "b", "a", "world"}
	resp, err := client.CalculateHashes(ctx, &hasherpb.HashRequest{Strings: in})
	require.NoError(t, err)
	require.NotNil(t, resp)

	hashes := resp.GetHashes()
	require.Len(t, hashes, len(in))
	require.Equal(t, hashes[0], hashes[2])
	require.NotEqual(t, hashes[0], hashes[1])
}

func TestCalculateHashes_Empty(t *testing.T) {
	conn, cleanup := startBufGRPC(t)
	defer cleanup()

	client := hasherpb.NewHasherServiceClient(conn)
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()

	resp, err := client.CalculateHashes(ctx, &hasherpb.HashRequest{Strings: []string{}})
	require.NoError(t, err)
	require.NotNil(t, resp)
	require.Len(t, resp.GetHashes(), 0)
}
