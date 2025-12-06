package rest

import (
	"net/http"
	"strconv"

	apperrors "github.com/emorenkov/scorehub/pkg/common/errors"
	"github.com/emorenkov/scorehub/pkg/notification"
	"github.com/labstack/echo/v4"
)

type createNotificationRequest struct {
	UserID  int64  `json:"user_id"`
	Message string `json:"message"`
}

type notificationDTO struct {
	ID        int64  `json:"id"`
	UserID    int64  `json:"user_id"`
	Message   string `json:"message"`
	CreatedAt string `json:"created_at"`
}

func (s *Server) createNotification(c echo.Context) error {
	var req createNotificationRequest
	if err := c.Bind(&req); err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "invalid json"})
	}
	n, err := s.svc.Create(c.Request().Context(), req.UserID, req.Message)
	if err != nil {
		if handled := writeServiceError(c, err); handled {
			return nil
		}
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": err.Error()})
	}
	return c.JSON(http.StatusCreated, toDTO(n))
}

func (s *Server) getNotification(c echo.Context) error {
	id, ok := parseID(c)
	if !ok {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "invalid id"})
	}
	n, err := s.svc.Get(c.Request().Context(), id)
	if err != nil {
		if handled := writeServiceError(c, err); handled {
			return nil
		}
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": err.Error()})
	}
	return c.JSON(http.StatusOK, toDTO(n))
}

func (s *Server) listNotifications(c echo.Context) error {
	userID := int64(0)
	if userIDStr := c.QueryParam("user_id"); userIDStr != "" {
		if val, err := strconv.ParseInt(userIDStr, 10, 64); err == nil {
			userID = val
		} else {
			return c.JSON(http.StatusBadRequest, map[string]string{"error": "invalid user_id"})
		}
	}
	notifications, err := s.svc.List(c.Request().Context(), userID)
	if err != nil {
		if handled := writeServiceError(c, err); handled {
			return nil
		}
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": err.Error()})
	}
	resp := make([]notificationDTO, 0, len(notifications))
	for i := range notifications {
		resp = append(resp, toDTO(&notifications[i]))
	}
	return c.JSON(http.StatusOK, resp)
}

func toDTO(n *notification.Notification) notificationDTO {
	return notificationDTO{
		ID:        n.ID,
		UserID:    n.UserID,
		Message:   n.Message,
		CreatedAt: n.CreatedAt.UTC().Format(timeRFC3339),
	}
}

func parseID(c echo.Context) (int64, bool) {
	idStr := c.Param("id")
	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		return 0, false
	}
	return id, true
}

func writeServiceError(c echo.Context, err error) bool {
	if err == nil {
		return false
	}
	if se, ok := apperrors.AsStatusError(err); ok {
		_ = c.JSON(se.Status, map[string]string{"error": se.Message})
		return true
	}
	return false
}

const timeRFC3339 = "2006-01-02T15:04:05Z07:00"
