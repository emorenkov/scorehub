package service

import (
	"context"
	"fmt"
	"net/http"
	"strings"
	"time"

	apperrors "github.com/emorenkov/scorehub/pkg/common/errors"
	"github.com/emorenkov/scorehub/pkg/notification"
	"github.com/emorenkov/scorehub/pkg/notification/producer"
	"github.com/emorenkov/scorehub/pkg/notification/repository"
	"gorm.io/gorm"
)

type Notification interface {
	Create(ctx context.Context, userID int64, message string) (*notification.Notification, error)
	Get(ctx context.Context, id int64) (*notification.Notification, error)
	List(ctx context.Context, userID int64) ([]notification.Notification, error)
	ProcessScoreEvent(ctx context.Context, ev *notification.ScoreEvent) (*notification.Notification, error)
}

type notificationService struct {
	repo      repository.Repository
	publisher producer.Publisher
}

func NewNotification(repo repository.Repository, publisher producer.Publisher) Notification {
	return &notificationService{repo: repo, publisher: publisher}
}

func (s *notificationService) Create(ctx context.Context, userID int64, message string) (*notification.Notification, error) {
	message = strings.TrimSpace(message)
	if userID <= 0 || message == "" {
		return nil, apperrors.NewStatusError(http.StatusBadRequest, "user_id and message are required")
	}
	n := &notification.Notification{
		UserID:  userID,
		Message: message,
	}
	if err := s.repo.Create(ctx, n); err != nil {
		return nil, apperrors.WrapStatus(err, http.StatusInternalServerError, "create notification")
	}
	if s.publisher != nil {
		_ = s.publisher.Publish(ctx, &notification.NotificationMessage{
			UserID:    n.UserID,
			Message:   n.Message,
			CreatedAt: time.Now().UTC(),
		})
	}
	return n, nil
}

func (s *notificationService) Get(ctx context.Context, id int64) (*notification.Notification, error) {
	if id <= 0 {
		return nil, apperrors.NewStatusError(http.StatusBadRequest, "invalid id")
	}
	n, err := s.repo.GetByID(ctx, id)
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, apperrors.NewStatusError(http.StatusNotFound, "notification not found")
		}
		return nil, apperrors.WrapStatus(err, http.StatusInternalServerError, "get notification")
	}
	return n, nil
}

func (s *notificationService) List(ctx context.Context, userID int64) ([]notification.Notification, error) {
	notifications, err := s.repo.List(ctx, userID)
	if err != nil {
		return nil, apperrors.WrapStatus(err, http.StatusInternalServerError, "list notifications")
	}
	return notifications, nil
}

func (s *notificationService) ProcessScoreEvent(ctx context.Context, ev *notification.ScoreEvent) (*notification.Notification, error) {
	if ev == nil {
		return nil, apperrors.NewStatusError(http.StatusBadRequest, "event is required")
	}
	if ev.Change <= 10 {
		return nil, nil
	}
	message := fmt.Sprintf("Congrats! Your score increased to %d (+%d)", ev.NewScore, ev.Change)
	return s.Create(ctx, ev.UserID, message)
}
