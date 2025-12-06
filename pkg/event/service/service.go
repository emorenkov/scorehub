package service

import (
	"context"
	"net/http"

	apperrors "github.com/emorenkov/scorehub/pkg/common/errors"
	"github.com/emorenkov/scorehub/pkg/event"
	"github.com/emorenkov/scorehub/pkg/event/repository"
)

type Service interface {
	Send(ctx context.Context, ev *event.ScoreEvent) (*event.EventAck, error)
}

type service struct {
	pub repository.Publisher
}

func NewService(pub repository.Publisher) Service {
	return &service{pub: pub}
}

func (s *service) Send(ctx context.Context, ev *event.ScoreEvent) (*event.EventAck, error) {
	if ev == nil {
		return nil, apperrors.NewStatusError(http.StatusBadRequest, "event is required")
	}
	if ev.UserID <= 0 {
		return nil, apperrors.NewStatusError(http.StatusBadRequest, "user_id must be positive")
	}
	if ev.NewScore < 0 {
		return nil, apperrors.NewStatusError(http.StatusBadRequest, "new_score must be non-negative")
	}

	if err := s.pub.PublishScoreEvent(ctx, ev); err != nil {
		return nil, apperrors.WrapStatus(err, http.StatusInternalServerError, "publish score event")
	}
	return &event.EventAck{Status: "ok"}, nil
}
