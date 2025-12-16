package app

import (
	"context"
	"fmt"
	"net"

	"github.com/emorenkov/scorehub/pkg/common/db"
	logpkg "github.com/emorenkov/scorehub/pkg/common/logger"
	"github.com/emorenkov/scorehub/pkg/user/config"
	grpcserver "github.com/emorenkov/scorehub/pkg/user/grpc"
	userpb "github.com/emorenkov/scorehub/pkg/user/models/proto"
	"github.com/emorenkov/scorehub/pkg/user/repository"
	"github.com/emorenkov/scorehub/pkg/user/rest"
	"github.com/emorenkov/scorehub/pkg/user/service"
	"go.uber.org/zap"
	"golang.org/x/sync/errgroup"
	"google.golang.org/grpc"
	"gorm.io/gorm"
)

// App wires together transport servers and shared dependencies for the user service.
type App struct {
	cfg          *config.UserConfig
	db           *gorm.DB
	restServer   *rest.Server
	grpcServer   *grpc.Server
	grpcListener net.Listener
}

// New constructs the application and its dependencies.
func New(cfg *config.UserConfig) (*App, error) {
	dbConn, err := db.NewPostgresDB(cfg.DbConfig)
	if err != nil {
		return nil, err
	}

	repo := repository.NewGormRepository(dbConn)
	svc := service.NewService(repo)

	restServer, err := rest.NewServer(cfg, svc, logpkg.Log)
	if err != nil {
		return nil, fmt.Errorf("init rest server: %w", err)
	}
	grpcServer := grpc.NewServer()
	userpb.RegisterUserServiceServer(grpcServer, grpcserver.NewServer(svc, logpkg.Log))
	grpcAddr := ":" + cfg.GRPCPort
	lis, err := net.Listen("tcp", grpcAddr)
	if err != nil {
		return nil, fmt.Errorf("failed to listen on %s: %w", grpcAddr, err)
	}

	return &App{
		cfg:          cfg,
		db:           dbConn,
		restServer:   restServer,
		grpcServer:   grpcServer,
		grpcListener: lis,
	}, nil
}

// Run starts HTTP and gRPC servers asynchronously and returns a channel of errors.
func (a *App) Run() <-chan error {
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

// Shutdown stops servers and closes shared resources gracefully.
func (a *App) Shutdown(ctx context.Context) error {
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
