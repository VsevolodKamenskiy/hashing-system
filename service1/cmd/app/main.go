package main

import (
	"context"
	"fmt"
	formatters "github.com/fabienm/go-logrus-formatters"
	graylog "github.com/gemnasium/logrus-graylog-hook/v3"
	"github.com/pkg/errors"
	"github.com/sirupsen/logrus"
	"google.golang.org/grpc"
	"net"
	"os"
	"os/signal"
	"service1/internal/server"
	"service1/proto/hasherpb"
	"syscall"
)

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

	grpcServer := grpc.NewServer(
		grpc.ChainUnaryInterceptor(server.UnaryServerLogger(log)),
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
