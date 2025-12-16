package rest

import (
	"net/http"

	apperrors "github.com/emorenkov/scorehub/pkg/common/errors"
	"github.com/emorenkov/scorehub/pkg/event"
	"github.com/labstack/echo/v4"
	"go.uber.org/zap"
)

type scoreEventRequest struct {
	UserID   int64 `json:"user_id"`
	NewScore int64 `json:"new_score"`
	Change   int32 `json:"change"`
}

func (s *Server) sendScoreEvent(c echo.Context) error {
	var req scoreEventRequest
	if err := c.Bind(&req); err != nil {
		s.log.Error("sendScoreEvent invalid json", zap.Error(err))
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "invalid json"})
	}

	ev := &event.ScoreEvent{
		UserID:   req.UserID,
		NewScore: req.NewScore,
		Change:   req.Change,
	}
	ack, err := s.svc.Send(c.Request().Context(), ev)
	if err != nil {
		s.log.Error("sendScoreEvent failed", zap.Error(err), zap.Int64("user_id", ev.UserID))
		if se, ok := apperrors.AsStatusError(err); ok {
			return c.JSON(se.Status, map[string]string{"error": se.Message})
		}
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": err.Error()})
	}
	s.log.Info("sendScoreEvent succeeded", zap.String("status", ack.Status), zap.Int64("user_id", ev.UserID))
	return c.JSON(http.StatusOK, ack)
}
