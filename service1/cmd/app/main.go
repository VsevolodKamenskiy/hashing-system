package main

import (
	"context"
	"fmt"
	formatters "github.com/fabienm/go-logrus-formatters"
	graylog "github.com/gemnasium/logrus-graylog-hook/v3"
	logging "github.com/grpc-ecosystem/go-grpc-middleware/v2/interceptors/logging"
	"github.com/pkg/errors"
	"github.com/sirupsen/logrus"
	"google.golang.org/grpc"
	"net"
	"os"
	"os/signal"
	"service1/internal/logctx"
	"service1/internal/server"
	"service1/proto/hasherpb"
	"syscall"
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

func main() {
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	log := logrus.New()
	logFormatter := formatters.NewGelf("service1")
	hook := graylog.NewGraylogHook("graylog:12201", map[string]interface{}{})

	log.Hooks.Add(hook)
	log.SetFormatter(logFormatter)

	listener, err := net.Listen("tcp", ":50051")
	if err != nil {
		werr := errors.WithStack(err)
		log.WithField("stack", fmt.Sprintf("%+v", werr)).WithError(werr).
			Error("failed to listen")
		return
	}

	logger := &logrusLogger{log}
	grpcServer := grpc.NewServer(
		grpc.ChainUnaryInterceptor(
			server.UnaryServerRequestID(log),
			logging.UnaryServerInterceptor(logger, logging.WithFieldsFromContext(func(ctx context.Context) logging.Fields {
				if v := ctx.Value(logctx.RequestIDKey); v != nil {
					if id, ok := v.(string); ok && id != "" {
						return logging.Fields{"request_id", id}
					}
				}
				return nil
			})),
		),
	)

	srv := &server.Server{
		Log:         log,
		ShutdownCtx: ctx,
	}

	hasherpb.RegisterHasherServiceServer(grpcServer, srv)

	go func() {
		log.Infoln("Starting Service 1 on port 50051...")
		if err := grpcServer.Serve(listener); err != nil {
			werr := errors.WithStack(err)
			log.WithField("stack", fmt.Sprintf("%+v", werr)).WithError(werr).
				Error("failed to serve")
		}
	}()

	<-ctx.Done()

	log.Infoln("Shutting down Service 1...")
	grpcServer.GracefulStop()
}
