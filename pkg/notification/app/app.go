package app

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/emorenkov/scorehub/pkg/common/db"
	ckafka "github.com/emorenkov/scorehub/pkg/common/kafka"
	logpkg "github.com/emorenkov/scorehub/pkg/common/logger"
	"github.com/emorenkov/scorehub/pkg/notification"
	"github.com/emorenkov/scorehub/pkg/notification/config"
	"github.com/emorenkov/scorehub/pkg/notification/producer"
	"github.com/emorenkov/scorehub/pkg/notification/repository"
	"github.com/emorenkov/scorehub/pkg/notification/rest"
	"github.com/emorenkov/scorehub/pkg/notification/service"
	"go.uber.org/zap"
	"golang.org/x/sync/errgroup"
	"gorm.io/gorm"
)

type App struct {
	cfg        *config.Config
	db         *gorm.DB
	restServer *rest.Server
	consumer   *ckafka.Consumer
	publisher  producer.Publisher
	svc        service.Notification
	cancel     context.CancelFunc
}

func New(cfg *config.Config) (*App, error) {
	dbConn, err := db.NewPostgresDB(cfg.DbConfig)
	if err != nil {
		return nil, fmt.Errorf("init db: %w", err)
	}
	if err := dbConn.AutoMigrate(&notification.Notification{}); err != nil {
		return nil, fmt.Errorf("migrate db: %w", err)
	}

	repo := repository.NewGormRepository(dbConn)
	pub := producer.NewKafkaPublisher(cfg.KafkaBrokers, cfg.NotificationsTopic)
	svc := service.NewNotification(repo, pub)

	restServer := rest.NewServer(cfg, svc, logpkg.Log)
	consumer := ckafka.NewConsumerWithBrokers(cfg.KafkaBrokers, cfg.ScoreEventsTopic, cfg.KafkaGroupID)

	return &App{
		cfg:        cfg,
		db:         dbConn,
		restServer: restServer,
		consumer:   consumer,
		publisher:  pub,
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
		logpkg.Log.Info("starting score events consumer", zap.String("topic", a.cfg.ScoreEventsTopic))
		for {
			msg, err := a.consumer.ReadMessage(ctx)
			if err != nil {
				if ctx.Err() != nil {
					return
				}
				errCh <- fmt.Errorf("consume score event: %w", err)
				return
			}
			var ev notification.ScoreEvent
			if err := json.Unmarshal(msg.Value, &ev); err != nil {
				logpkg.Log.Error("failed to unmarshal score event", zap.Error(err))
				continue
			}
			if _, err := a.svc.ProcessScoreEvent(ctx, &ev); err != nil {
				logpkg.Log.Error("process score event failed", zap.Error(err))
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

	g.Go(func() error {
		if a.publisher != nil {
			return a.publisher.Close()
		}
		return nil
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
