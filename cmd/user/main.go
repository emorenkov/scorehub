package main

import (
	"context"
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
	"github.com/emorenkov/scorehub/pkg/user/rest"
	"github.com/emorenkov/scorehub/pkg/user/service"
	"go.uber.org/zap"
	"golang.org/x/sync/errgroup"
	"google.golang.org/grpc"
	"gorm.io/gorm"
)

func main() {
	cfg := usercfg.Load()

	if err := logpkg.Init(cfg.ServiceName); err != nil {
		panic(fmt.Errorf("failed to init logger: %w", err))
	}
	defer logpkg.Sync()

	application, err := newApp(cfg)
	if err != nil {
		logpkg.Log.Fatal("failed to init app", zap.Error(err))
	}

	errCh := application.run()
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

	if err := application.shutdown(ctx); err != nil {
		logpkg.Log.Error("graceful shutdown failed", zap.Error(err))
	}
}

type app struct {
	cfg          *usercfg.UserConfig
	db           *gorm.DB
	restServer   *rest.Server
	grpcServer   *grpc.Server
	grpcListener net.Listener
}

func newApp(cfg *usercfg.UserConfig) (*app, error) {
	dbConn, err := db.NewPostgresDB(cfg.DbConfig)
	if err != nil {
		return nil, err
	}

	repo := repository.NewGormRepository(dbConn)
	svc := service.NewService(repo)

	restServer := rest.NewServer(cfg, svc, logpkg.Log)
	grpcServer := grpc.NewServer()
	grpcAddr := ":" + cfg.GRPCPort
	lis, err := net.Listen("tcp", grpcAddr)
	if err != nil {
		return nil, fmt.Errorf("failed to listen on %s: %w", grpcAddr, err)
	}

	return &app{
		cfg:          cfg,
		db:           dbConn,
		restServer:   restServer,
		grpcServer:   grpcServer,
		grpcListener: lis,
	}, nil
}

func (a *app) run() <-chan error {
	errCh := make(chan error, 2)

	go func() {
		if err := a.restServer.Serve(); err != nil {
			errCh <- fmt.Errorf("rest server: %w", err)
		}
	}()

	go func() {
		logpkg.Log.Info("starting gRPC server", zap.String("addr", ":"+a.cfg.GRPCPort))
		if err := a.grpcServer.Serve(a.grpcListener); err != nil {
			errCh <- fmt.Errorf("grpc server: %w", err)
		}
	}()

	return errCh
}

func (a *app) shutdown(ctx context.Context) error {
	g, ctx := errgroup.WithContext(ctx)

	g.Go(func() error {
		return a.restServer.Shutdown(ctx)
	})

	g.Go(func() error {
		done := make(chan struct{})
		go func() {
			a.grpcServer.GracefulStop()
			close(done)
		}()

		select {
		case <-done:
			return nil
		case <-ctx.Done():
			a.grpcServer.Stop()
			return ctx.Err()
		}
	})

	g.Go(func() error {
		sqlDB, err := a.db.DB()
		if err != nil {
			return err
		}
		return sqlDB.Close()
	})

	return g.Wait()
}
