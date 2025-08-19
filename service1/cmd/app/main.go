package main

import (
	"context"
	"fmt"
	formatters "github.com/fabienm/go-logrus-formatters"
	graylog "github.com/gemnasium/logrus-graylog-hook/v3"
	"github.com/pkg/errors"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"github.com/sirupsen/logrus"
	"google.golang.org/grpc"
	"net"
	"net/http"
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

	grpcMetrics := server.NewServerMetrics()
	prometheus.MustRegister(grpcMetrics)

	grpcServer := grpc.NewServer(
		grpc.ChainUnaryInterceptor(
			server.UnaryRequestID(log),
			server.LoggingInterceptor(log),
			grpcMetrics.UnaryServerInterceptor(),
		),
	)

	srv := &server.Server{
		Log:         log,
		ShutdownCtx: ctx,
	}

	hasherpb.RegisterHasherServiceServer(grpcServer, srv)
	grpcMetrics.InitializeMetrics(grpcServer)

	go func() {
		mux := http.NewServeMux()
		mux.Handle("/metrics", promhttp.Handler())
		if err := http.ListenAndServe(":2112", mux); err != nil && !errors.Is(err, http.ErrServerClosed) {
			werr := errors.WithStack(err)
			log.WithField("stack", fmt.Sprintf("%+v", werr)).WithError(werr).Error("metrics server failed")
		}
	}()

	log.Infoln("Starting Service 1 on port 50051...")
	go func() {

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
