package service

import (
	"context"
	"net/http"
	"strings"

	apperrors "github.com/emorenkov/scorehub/pkg/common/errors"
	"github.com/emorenkov/scorehub/pkg/email/repository"
)

type Email interface {
	Send(ctx context.Context, userID int64, message string) error
}

type email struct {
	sender repository.Sender
}

func NewEmail(sender repository.Sender) Email {
	return &email{sender: sender}
}

func (s *email) Send(ctx context.Context, userID int64, message string) error {
	message = strings.TrimSpace(message)
	if userID <= 0 || message == "" {
		return apperrors.NewStatusError(http.StatusBadRequest, "user_id and message are required")
	}
	if err := s.sender.Send(ctx, userID, message); err != nil {
		return apperrors.WrapStatus(err, http.StatusInternalServerError, "send email")
	}
	return nil
}
