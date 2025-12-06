package rest

import (
	"net/http"

	apperrors "github.com/emorenkov/scorehub/pkg/common/errors"
	"github.com/emorenkov/scorehub/pkg/event"
	"github.com/labstack/echo/v4"
)

type scoreEventRequest struct {
	UserID   int64 `json:"user_id"`
	NewScore int64 `json:"new_score"`
	Change   int32 `json:"change"`
}

func (s *Server) sendScoreEvent(c echo.Context) error {
	var req scoreEventRequest
	if err := c.Bind(&req); err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "invalid json"})
	}

	ev := &event.ScoreEvent{
		UserID:   req.UserID,
		NewScore: req.NewScore,
		Change:   req.Change,
	}
	ack, err := s.svc.Send(c.Request().Context(), ev)
	if err != nil {
		if se, ok := apperrors.AsStatusError(err); ok {
			return c.JSON(se.Status, map[string]string{"error": se.Message})
		}
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": err.Error()})
	}
	return c.JSON(http.StatusOK, ack)
}
