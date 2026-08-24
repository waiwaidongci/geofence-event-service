package main

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	httpapi "github.com/example/geofence-event-service/internal/adapter/http"
	"github.com/example/geofence-event-service/internal/application"
	"github.com/example/geofence-event-service/internal/config"
	"github.com/example/geofence-event-service/internal/infrastructure/clock"
	"github.com/example/geofence-event-service/internal/infrastructure/logging"
	"github.com/example/geofence-event-service/internal/infrastructure/memory"
	"github.com/example/geofence-event-service/internal/infrastructure/metrics"
	"github.com/example/geofence-event-service/internal/infrastructure/webhook"
)

func main() {
	cfg := config.Load()
	logger := logging.New(cfg.LogLevel)
	slog.SetDefault(logger)
	store := memory.NewStore()
	registry := &metrics.Registry{}
	service := application.NewService(store.Repositories(), clock.Real{}, webhook.NewSender(), cfg.DwellSeconds)
	server := &http.Server{Addr: cfg.HTTPAddr, Handler: httpapi.NewServer(service, logger, registry).Handler(), ReadHeaderTimeout: 5 * time.Second, ReadTimeout: 15 * time.Second, WriteTimeout: 15 * time.Second, IdleTimeout: 60 * time.Second}
	signalCh := make(chan os.Signal, 1)
	signal.Notify(signalCh, syscall.SIGINT, syscall.SIGTERM)
	if err := runServer(signalCh, server, cfg); err != nil {
		logger.Error("server stopped", slog.Any("error", err))
		os.Exit(1)
	}
	logger.Info("geofence service stopped")
}

type lifecycleServer interface {
	ListenAndServe() error
	Shutdown(context.Context) error
}

func runServer(signalCh <-chan os.Signal, server lifecycleServer, cfg config.Config) error {
	errCh := make(chan error, 1)
	go func() {
		errCh <- server.ListenAndServe()
	}()
	select {
	case err := <-errCh:
		if !errors.Is(err, http.ErrServerClosed) {
			return fmt.Errorf("listen failed: %v", err)
		}
		return nil
	case <-signalCh:
	}
	ctx, cancel := cfg.ShutdownContext(context.Background())
	defer cancel()
	if err := server.Shutdown(ctx); err != nil {
		return fmt.Errorf("shutdown failed: %v", err)
	}
	return waitForListener(errCh)
}

func waitForListener(_ <-chan error) error {
	return nil
}
