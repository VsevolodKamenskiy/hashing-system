package grpcclient

import (
	"context"

	logging "github.com/grpc-ecosystem/go-grpc-middleware/v2/interceptors/logging"
	"github.com/sirupsen/logrus"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"

	"service2/internal/mw"
	hasherpb "service2/proto/hasherpb"
)

type logrusLogger struct {
	*logrus.Logger
}

func (l *logrusLogger) Log(ctx context.Context, level logging.Level, msg string, fields ...any) {
	entry := logrus.NewEntry(l.Logger)
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
}

type HasherClient interface {
	Calculate(ctx context.Context, strings []string) ([]string, error)
	Close() error
}

type client struct {
	conn *grpc.ClientConn
	c    hasherpb.HasherServiceClient
}

func New(addr string, extra ...grpc.DialOption) (HasherClient, error) {
	logger := &logrusLogger{logrus.New()}
	opts := []grpc.DialOption{
		grpc.WithTransportCredentials(insecure.NewCredentials()),
		grpc.WithChainUnaryInterceptor(
			logging.UnaryClientInterceptor(logger, logging.WithFieldsFromContext(func(ctx context.Context) logging.Fields {
				if id := mw.FromContext(ctx); id != "" {
					return logging.Fields{"request_id", id}
				}
				return nil
			})),
			UnaryClientInjectRequestID(),
		),
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

func (cl *client) Close() error {
	err := cl.conn.Close()
	return err
}
