package main

import (
	"fmt"
	"net"
	"os"
	"os/signal"
	"syscall"

	logpkg "github.com/emorenkov/scorehub/pkg/common/logger"
	eventcfg "github.com/emorenkov/scorehub/pkg/event/config"
	"go.uber.org/zap"
	"google.golang.org/grpc"
)

func main() {
	cfg := eventcfg.Load()

	if err := logpkg.Init(cfg.ServiceName); err != nil {
		panic(fmt.Errorf("failed to init logger: %w", err))
	}
	defer logpkg.Sync()

	addr := ":" + cfg.GRPCPort
	lis, err := net.Listen("tcp", addr)
	if err != nil {
		logpkg.Log.Fatal("failed to listen", zap.Error(err), zap.String("addr", addr))
	}

	grpcServer := grpc.NewServer()
	// TODO: register event service handler once implemented and protobuf generated

	go func() {
		logpkg.Log.Info("starting gRPC server", zap.String("addr", addr))
		if err := grpcServer.Serve(lis); err != nil {
			logpkg.Log.Fatal("gRPC server exited", zap.Error(err))
		}
	}()

	sigs := make(chan os.Signal, 1)
	signal.Notify(sigs, syscall.SIGINT, syscall.SIGTERM)
	<-sigs
	logpkg.Log.Info("shutting down gRPC server")
	grpcServer.GracefulStop()
}
