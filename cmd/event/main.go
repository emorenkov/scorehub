package main

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"syscall"
	"time"

	logpkg "github.com/emorenkov/scorehub/pkg/common/logger"
	"github.com/emorenkov/scorehub/pkg/event/app"
	eventcfg "github.com/emorenkov/scorehub/pkg/event/config"
	"go.uber.org/zap"
)

func main() {
	cfg := eventcfg.Load()

	if err := logpkg.Init(cfg.ServiceName); err != nil {
		panic(fmt.Errorf("failed to init logger: %w", err))
	}
	defer logpkg.Sync()

	application, err := app.New(cfg)
	if err != nil {
		logpkg.Log.Fatal("failed to init app", zap.Error(err))
	}

	errCh := application.Run()

	sigs := make(chan os.Signal, 1)
	signal.Notify(sigs, syscall.SIGINT, syscall.SIGTERM)

	select {
	case sig := <-sigs:
		logpkg.Log.Info("shutdown signal received", zap.String("signal", sig.String()))
	case err = <-errCh:
		if err != nil {
			logpkg.Log.Error("service error", zap.Error(err))
		}
	}

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	if err := application.Shutdown(ctx); err != nil {
		logpkg.Log.Error("graceful shutdown failed", zap.Error(err))
	}
}
