package main

import (
	"fmt"
	"os"
	"os/signal"
	"syscall"

	logpkg "github.com/emorenkov/scorehub/pkg/common/logger"
	emailcfg "github.com/emorenkov/scorehub/pkg/email/config"
)

func main() {
	cfg := emailcfg.Load()

	if err := logpkg.Init(cfg.ServiceName); err != nil {
		panic(fmt.Errorf("failed to init logger: %w", err))
	}
	defer logpkg.Sync()

	logpkg.Log.Info("email-service started; waiting for shutdown signal")

	sigs := make(chan os.Signal, 1)
	signal.Notify(sigs, syscall.SIGINT, syscall.SIGTERM)
	<-sigs
	logpkg.Log.Info("email-service shutting down")
}
