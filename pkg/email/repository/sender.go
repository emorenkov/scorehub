package repository

import (
	"context"

	"go.uber.org/zap"
)

type Sender interface {
	Send(ctx context.Context, userID int64, message string) error
}

// LoggerSender simulates email delivery by logging the payload.
type LoggerSender struct {
	log *zap.Logger
}

func NewLoggerSender(log *zap.Logger) *LoggerSender {
	return &LoggerSender{log: log}
}

func (s *LoggerSender) Send(_ context.Context, userID int64, message string) error {
	s.log.Info("sending email", zap.Int64("user_id", userID), zap.String("message", message))
	return nil
}
