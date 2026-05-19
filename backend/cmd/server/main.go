package main

import (
	"context"
	"fmt"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/your-org/itsm_ops/backend/internal/alert"
	"github.com/your-org/itsm_ops/backend/internal/api"
	"github.com/your-org/itsm_ops/backend/internal/config"
	"github.com/your-org/itsm_ops/backend/internal/notify"
	"github.com/your-org/itsm_ops/backend/pkg/db"
	"github.com/your-org/itsm_ops/backend/pkg/logger"
)

func main() {
	cfg := config.Load()
	log := logger.New(cfg.App.LogLevel)
	slog.SetDefault(log)

	conn, err := db.Open(db.Config{
		Driver:   cfg.DB.Type,
		Host:     cfg.DB.Host,
		Port:     cfg.DB.Port,
		Name:     cfg.DB.Name,
		User:     cfg.DB.User,
		Password: cfg.DB.Password,
	})
	if err != nil {
		log.Error("database connection failed", "error", err)
		os.Exit(1)
	}
	defer conn.Close()
	log.Info("database connected", "driver", cfg.DB.Type)

	server := &http.Server{
		Addr:         fmt.Sprintf(":%d", cfg.HTTP.Port),
		Handler:      api.NewRouter(api.RouterOptions{Config: cfg, Logger: log, DB: conn}),
		ReadTimeout:  cfg.HTTP.ReadTimeout,
		WriteTimeout: cfg.HTTP.WriteTimeout,
	}

	stopCh := make(chan os.Signal, 1)
	signal.Notify(stopCh, syscall.SIGINT, syscall.SIGTERM)

	alertTicker := time.NewTicker(30 * time.Second)
	defer alertTicker.Stop()
	go func() {
		for range alertTicker.C {
			alert.Evaluate(conn)
			notify.SendNotifications(conn)
		}
	}()
	log.Info("alert engine started", "interval", "30s")

	errCh := make(chan error, 1)
	go func() {
		log.Info("backend server starting", "addr", server.Addr, "env", cfg.App.Env)
		errCh <- server.ListenAndServe()
	}()

	select {
	case sig := <-stopCh:
		log.Info("shutdown signal received", "signal", sig.String())
		ctx, cancel := context.WithTimeout(context.Background(), cfg.HTTP.ShutdownTimeout)
		defer cancel()
		if err := server.Shutdown(ctx); err != nil {
			log.Error("server shutdown failed", "error", err)
			os.Exit(1)
		}
	case err := <-errCh:
		if err != nil && err != http.ErrServerClosed {
			log.Error("server failed", "error", err)
			os.Exit(1)
		}
	}
	log.Info("backend server stopped")
}
