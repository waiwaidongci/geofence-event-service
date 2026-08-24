package main

import (
	"context"
	"errors"
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
	errCh := make(chan error, 1)
	go func() {
		logger.Info("geofence service started", slog.String("addr", cfg.HTTPAddr))
		errCh <- server.ListenAndServe()
	}()
	signalCh := make(chan os.Signal, 1)
	signal.Notify(signalCh, syscall.SIGINT, syscall.SIGTERM)
	select {
	case err := <-errCh:
		if !errors.Is(err, http.ErrServerClosed) {
			logger.Error("server stopped", slog.Any("error", err))
			os.Exit(1)
		}
	case sig := <-signalCh:
		logger.Info("shutdown requested", slog.String("signal", sig.String()))
	}
	ctx, cancel := context.WithTimeout(context.Background(), time.Duration(cfg.ShutdownSeconds)*time.Second)
	defer cancel()
	if err := server.Shutdown(ctx); err != nil {
		logger.Error("graceful shutdown failed", slog.Any("error", err))
		os.Exit(1)
	}
	logger.Info("geofence service stopped")
}
