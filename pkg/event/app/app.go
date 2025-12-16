package app

import (
	"context"
	"fmt"
	"net"

	logpkg "github.com/emorenkov/scorehub/pkg/common/logger"
	"github.com/emorenkov/scorehub/pkg/event/config"
	grpcserver "github.com/emorenkov/scorehub/pkg/event/grpc"
	eventpb "github.com/emorenkov/scorehub/pkg/event/proto"
	"github.com/emorenkov/scorehub/pkg/event/repository"
	"github.com/emorenkov/scorehub/pkg/event/rest"
	"github.com/emorenkov/scorehub/pkg/event/service"
	userpb "github.com/emorenkov/scorehub/pkg/user/models/proto"
	"go.uber.org/zap"
	"golang.org/x/sync/errgroup"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

type App struct {
	cfg          *config.Config
	restServer   *rest.Server
	grpcServer   *grpc.Server
	grpcListener net.Listener
	publisher    repository.Publisher
	userConn     *grpc.ClientConn
}

func New(cfg *config.Config) (*App, error) {
	pub, err := repository.NewKafkaPublisher(cfg.KafkaBrokers, cfg.ScoreEventsTopic)
	if err != nil {
		return nil, fmt.Errorf("init kafka publisher: %w", err)
	}

	userConn, err := grpc.Dial(cfg.UserServiceAddr, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		return nil, fmt.Errorf("dial user service: %w", err)
	}

	svc := service.NewEvent(pub, userpb.NewUserServiceClient(userConn))

	restServer := rest.NewServer(cfg, svc, logpkg.Log)

	grpcSrv := grpc.NewServer()
	eventpb.RegisterEventServiceServer(grpcSrv, grpcserver.NewServer(svc, logpkg.Log))
	grpcAddr := ":" + cfg.GRPCPort
	lis, err := net.Listen("tcp", grpcAddr)
	if err != nil {
		return nil, fmt.Errorf("failed to listen on %s: %w", grpcAddr, err)
	}

	return &App{
		cfg:          cfg,
		restServer:   restServer,
		grpcServer:   grpcSrv,
		grpcListener: lis,
		publisher:    pub,
		userConn:     userConn,
	}, nil
}

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
		return a.publisher.Close()
	})

	g.Go(func() error {
		if a.userConn != nil {
			return a.userConn.Close()
		}
		return nil
	})

	return g.Wait()
}
