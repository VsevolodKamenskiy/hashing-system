package main

import (
	"context"
	"fmt"
	"github.com/pkg/errors"
	"net/http"
	"os"
	"os/signal"
	"service2/internal/config"
	"syscall"
	"time"

	"service2/internal/api"
	"service2/internal/grpcclient"
	"service2/internal/mw"
	"service2/internal/storage"
)

func main() {

	// базовый контекст с отменой по сигналам
	rootCtx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	logg := mw.NewLogger("service2")

	appCfg, err := config.Load("consul:8500")
	if err != nil {
		werr := errors.WithStack(err)
		logg.WithField("stack", fmt.Sprintf("%+v", werr)).WithError(err).
			Error("consul init error")
		return
	}

	store, err := storage.New(rootCtx, appCfg.DBDSN)
	if err != nil {
		werr := errors.WithStack(err)
		logg.WithField("stack", fmt.Sprintf("%+v", werr)).WithError(werr).
			Error("db connect failed")
		return
	}
	defer store.Close()

	hashCl, err := grpcclient.New(fmt.Sprintf("service1:%s", appCfg.HasherPort))
	defer hashCl.Close()

	if err != nil {
		werr := errors.WithStack(err)
		logg.WithField("stack", fmt.Sprintf("%+v", werr)).WithError(werr).
			Fatal("grpc client failed")
	}

	h := &api.Handlers{HashClient: hashCl, Store: store, Log: logg}
	r := api.NewRouter(h, logg)

	httpAddr := fmt.Sprintf(":%s", appCfg.HTTPPort)

	srv := &http.Server{
		Addr:         httpAddr,
		Handler:      r,
		ReadTimeout:  10 * time.Second,
		WriteTimeout: 30 * time.Second,
		IdleTimeout:  60 * time.Second,
	}

	go func() {
		logg.Printf("service2 listening on %s", httpAddr)
		if err := srv.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			werr := errors.WithStack(err)
			logg.WithField("stack", fmt.Sprintf("%+v", werr)).WithError(werr).
				Fatal("http server failed")
		}
	}()

	<-rootCtx.Done()
	logg.Info("service2: shutting down...")

	// даём 10 секунд на graceful
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	_ = srv.Shutdown(ctx)
}
