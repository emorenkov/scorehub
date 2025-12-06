package app

import (
	"context"
	"encoding/json"
	"fmt"

	ckafka "github.com/emorenkov/scorehub/pkg/common/kafka"
	logpkg "github.com/emorenkov/scorehub/pkg/common/logger"
	"github.com/emorenkov/scorehub/pkg/email/config"
	"github.com/emorenkov/scorehub/pkg/email/repository"
	"github.com/emorenkov/scorehub/pkg/email/rest"
	"github.com/emorenkov/scorehub/pkg/email/service"
	"github.com/emorenkov/scorehub/pkg/notification"
	"go.uber.org/zap"
	"golang.org/x/sync/errgroup"
)

type App struct {
	cfg        *config.Config
	restServer *rest.Server
	consumer   *ckafka.Consumer
	svc        service.Service
	cancel     context.CancelFunc
}

func New(cfg *config.Config) (*App, error) {
	sender := repository.NewLoggerSender(logpkg.Log)
	svc := service.NewService(sender)
	restServer := rest.NewServer(cfg, svc, logpkg.Log)
	consumer := ckafka.NewConsumerWithBrokers(cfg.KafkaBrokers, cfg.NotificationsTopic, cfg.KafkaGroupID)

	return &App{
		cfg:        cfg,
		restServer: restServer,
		consumer:   consumer,
		svc:        svc,
	}, nil
}

func (a *App) Run() <-chan error {
	errCh := make(chan error, 2)
	ctx, cancel := context.WithCancel(context.Background())
	a.cancel = cancel

	go func() {
		if err := a.restServer.Serve(); err != nil {
			errCh <- fmt.Errorf("rest server: %w", err)
		}
	}()

	go func() {
		logpkg.Log.Info("starting notifications consumer", zap.String("topic", a.cfg.NotificationsTopic))
		for {
			msg, err := a.consumer.ReadMessage(ctx)
			if err != nil {
				if ctx.Err() != nil {
					return
				}
				errCh <- fmt.Errorf("consume notification: %w", err)
				return
			}
			var notif notification.NotificationMessage
			if err := json.Unmarshal(msg.Value, &notif); err != nil {
				logpkg.Log.Error("failed to unmarshal notification", zap.Error(err))
				continue
			}
			if err := a.svc.Send(ctx, notif.UserID, notif.Message); err != nil {
				logpkg.Log.Error("failed to send email", zap.Error(err))
			}
		}
	}()

	return errCh
}

func (a *App) Shutdown(ctx context.Context) error {
	if a.cancel != nil {
		a.cancel()
	}

	g, ctx := errgroup.WithContext(ctx)

	g.Go(func() error {
		return a.restServer.Shutdown(ctx)
	})

	g.Go(func() error {
		return a.consumer.Close()
	})

	return g.Wait()
}
