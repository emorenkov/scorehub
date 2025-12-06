package rest

import (
	"net/http"

	apperrors "github.com/emorenkov/scorehub/pkg/common/errors"
	"github.com/labstack/echo/v4"
)

type sendEmailRequest struct {
	UserID  int64  `json:"user_id"`
	Message string `json:"message"`
}

type sendEmailResponse struct {
	Status string `json:"status"`
}

func (s *Server) sendEmail(c echo.Context) error {
	var req sendEmailRequest
	if err := c.Bind(&req); err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "invalid json"})
	}

	if err := s.svc.Send(c.Request().Context(), req.UserID, req.Message); err != nil {
		if se, ok := apperrors.AsStatusError(err); ok {
			return c.JSON(se.Status, map[string]string{"error": se.Message})
		}
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": err.Error()})
	}

	return c.JSON(http.StatusOK, sendEmailResponse{Status: "sent"})
}
