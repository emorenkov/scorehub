package main

import (
	"fmt"
	"net"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/emorenkov/scorehub/pkg/common/db"
	logpkg "github.com/emorenkov/scorehub/pkg/common/logger"
	usercfg "github.com/emorenkov/scorehub/pkg/user/config"
	"github.com/emorenkov/scorehub/pkg/user/repository"
	userserver "github.com/emorenkov/scorehub/pkg/user/server"
	"github.com/emorenkov/scorehub/pkg/user/service"
	"go.uber.org/zap"
	"google.golang.org/grpc"
)

func main() {
	cfg := usercfg.Load()

	if err := logpkg.Init(cfg.ServiceName); err != nil {
		panic(fmt.Errorf("failed to init logger: %w", err))
	}
	defer logpkg.Sync()

	// Initialize database
	dbConn, err := db.NewPostgresDB(cfg.DbConfig)
	if err != nil {
		logpkg.Log.Fatal("failed to init postgres", zap.Error(err))
	}

	repo := repository.NewGormRepository(dbConn)
	svc := service.NewService(repo)

	grpcServer := grpc.NewServer()

	// Shared shutdown channel
	done := make(chan struct{})

	// Start gRPC
	go func() {
		grpcAddr := ":" + cfg.GRPCPort
		lis, err := net.Listen("tcp", grpcAddr)
		if err != nil {
			logpkg.Log.Fatal("failed to listen", zap.Error(err), zap.String("addr", grpcAddr))
		}
		logpkg.Log.Info("starting gRPC server", zap.String("addr", grpcAddr))
		if err := grpcServer.Serve(lis); err != nil {
			logpkg.Log.Fatal("gRPC server exited", zap.Error(err))
		}
	}()

	srv := userserver.NewServer(svc, logpkg.Log, cfg)
	// Start native REST (fasthttp)
	srv.StartREST(done)

	// graceful shutdown
	sigs := make(chan os.Signal, 1)
	signal.Notify(sigs, syscall.SIGINT, syscall.SIGTERM)
	<-sigs
	logpkg.Log.Info("shutting down servers")
	close(done)

	// stop gRPC
	stopCh := make(chan struct{})
	go func() {
		grpcServer.GracefulStop()
		close(stopCh)
	}()

	select {
	case <-stopCh:
	case <-time.After(5 * time.Second):
		logpkg.Log.Warn("timeout waiting for gRPC shutdown")
	}
}
