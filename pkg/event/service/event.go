package service

import (
	"context"
	"net/http"

	apperrors "github.com/emorenkov/scorehub/pkg/common/errors"
	"github.com/emorenkov/scorehub/pkg/event"
	"github.com/emorenkov/scorehub/pkg/event/repository"
	userpb "github.com/emorenkov/scorehub/pkg/user/models/proto"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type Event interface {
	Send(ctx context.Context, ev *event.ScoreEvent) (*event.EventAck, error)
}

type eventService struct {
	pub        repository.Publisher
	userClient userpb.UserServiceClient
}

func NewEvent(pub repository.Publisher, userClient userpb.UserServiceClient) Event {
	return &eventService{pub: pub, userClient: userClient}
}

func (s *eventService) Send(ctx context.Context, ev *event.ScoreEvent) (*event.EventAck, error) {
	if ev == nil {
		return nil, apperrors.NewStatusError(http.StatusBadRequest, "event is required")
	}
	if ev.UserID <= 0 {
		return nil, apperrors.NewStatusError(http.StatusBadRequest, "user_id must be positive")
	}
	if ev.NewScore < 0 {
		return nil, apperrors.NewStatusError(http.StatusBadRequest, "new_score must be non-negative")
	}

	if s.userClient == nil {
		return nil, apperrors.NewStatusError(http.StatusInternalServerError, "user client not configured")
	}

	if _, err := s.userClient.GetUser(ctx, &userpb.GetUserRequest{Id: ev.UserID}); err != nil {
		if status.Code(err) == codes.NotFound {
			return nil, apperrors.NewStatusError(http.StatusNotFound, "not found")
		}
		return nil, apperrors.WrapStatus(err, http.StatusInternalServerError, "validate user")
	}

	if err := s.pub.PublishScoreEvent(ctx, ev); err != nil {
		return nil, apperrors.WrapStatus(err, http.StatusInternalServerError, "publish score event")
	}
	return &event.EventAck{Status: "ok"}, nil
}
