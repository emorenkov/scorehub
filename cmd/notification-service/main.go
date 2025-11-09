package main

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/emorenkov/scorehub/pkg/common/db"
	ckafka "github.com/emorenkov/scorehub/pkg/common/kafka"
	logpkg "github.com/emorenkov/scorehub/pkg/common/logger"
	notif "github.com/emorenkov/scorehub/pkg/notification"
	notifcfg "github.com/emorenkov/scorehub/pkg/notification/config"
	"go.uber.org/zap"
)

func main() {
	cfg := notifcfg.Load()

	if err := logpkg.Init(cfg.ServiceName); err != nil {
		panic(fmt.Errorf("failed to init logger: %w", err))
	}
	defer logpkg.Sync()

	// Init DB and migrate
	dbConn, err := db.NewPostgresDBWithParams(cfg.PostgresHost, cfg.PostgresPort, cfg.PostgresUser, cfg.PostgresPassword, cfg.PostgresDB)
	if err != nil {
		logpkg.Log.Fatal("db init failed", zap.Error(err))
	}
	if err := dbConn.AutoMigrate(&notif.Notification{}); err != nil {
		logpkg.Log.Fatal("db migrate failed", zap.Error(err))
	}
	repo := notif.NewGormRepository(dbConn)

	consumer := ckafka.NewConsumerWithBrokers(cfg.KafkaBrokers, cfg.ScoreEventsTopic, cfg.KafkaGroupID)
	producer := ckafka.NewProducerWithBrokers(cfg.KafkaBrokers, cfg.NotificationsTopic)
	defer consumer.Close()
	defer producer.Close()

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	logpkg.Log.Info("notification-service started; consuming score events", zap.String("topic", cfg.ScoreEventsTopic))

	sigs := make(chan os.Signal, 1)
	signal.Notify(sigs, syscall.SIGINT, syscall.SIGTERM)

	go func() {
		for {
			msg, err := consumer.ReadMessage(ctx)
			if err != nil {
				logpkg.Log.Error("consumer read error", zap.Error(err))
				// brief backoff
				time.Sleep(500 * time.Millisecond)
				continue
			}
			var ev notif.ScoreEvent
			if err := json.Unmarshal(msg.Value, &ev); err != nil {
				logpkg.Log.Error("failed to unmarshal score event", zap.Error(err))
				continue
			}

			// Business rule: if increase > 10 -> persist notification and publish
			if ev.Change > 10 {
				message := fmt.Sprintf("Congrats! Your score increased to %d (+%d)", ev.NewScore, ev.Change)
				n := &notif.Notification{UserID: ev.UserID, Message: message}
				if err := repo.Create(ctx, n); err != nil {
					logpkg.Log.Error("failed to insert notification", zap.Error(err))
					continue
				}
				// publish to notifications topic
				_ = producer.SendMessage(ctx, fmt.Sprintf("user:%d", ev.UserID), map[string]interface{}{
					"user_id":    ev.UserID,
					"message":    message,
					"created_at": time.Now().Format(time.RFC3339),
				})
				logpkg.Log.Info("notification created", zap.Int64("user_id", ev.UserID))
			} else {
				logpkg.Log.Info("event below threshold; skipping notification", zap.Int32("change", ev.Change))
			}
		}
	}()

	<-sigs
	logpkg.Log.Info("notification-service shutting down")
	cancel()
}
