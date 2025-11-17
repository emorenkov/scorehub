package rest

import (
	"encoding/json"
	"net/http"
	"strconv"

	apperrors "github.com/emorenkov/scorehub/pkg/common/errors"
	"github.com/emorenkov/scorehub/pkg/common/models"
	"github.com/labstack/echo/v4"
)

func (s *Server) createUser(c echo.Context) error {
	var req createUserRequest
	if err := json.NewDecoder(c.Request().Body).Decode(&req); err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "invalid json"})
	}
	u, err := s.svc.Create(c.Request().Context(), req.Name, req.Email)
	if err != nil {
		if handled := writeServiceError(c, err); handled {
			return nil
		}
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": err.Error()})
	}
	return c.JSON(http.StatusCreated, toDTO(u))
}

func (s *Server) getUser(c echo.Context) error {
	id, ok := parseID(c)
	if !ok {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "invalid id"})
	}
	u, err := s.svc.Get(c.Request().Context(), id)
	if err != nil {
		if handled := writeServiceError(c, err); handled {
			return nil
		}
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": err.Error()})
	}
	return c.JSON(http.StatusOK, toDTO(u))
}

func (s *Server) listUsers(c echo.Context) error {
	users, err := s.svc.List(c.Request().Context())
	if err != nil {
		if handled := writeServiceError(c, err); handled {
			return nil
		}
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": err.Error()})
	}
	resp := make([]userDTO, 0, len(users))
	for i := range users {
		resp = append(resp, toDTO(&users[i]))
	}
	return c.JSON(http.StatusOK, resp)
}

func (s *Server) updateUser(c echo.Context) error {
	id, ok := parseID(c)
	if !ok {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "invalid id"})
	}
	var req updateUserRequest
	if err := json.NewDecoder(c.Request().Body).Decode(&req); err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "invalid json"})
	}

	u, err := s.svc.Update(c.Request().Context(), id, req.Name, req.Email)
	if err != nil {
		if handled := writeServiceError(c, err); handled {
			return nil
		}
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": err.Error()})
	}
	return c.JSON(http.StatusOK, toDTO(u))
}

func (s *Server) deleteUser(c echo.Context) error {
	id, ok := parseID(c)
	if !ok {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "invalid id"})
	}
	if err := s.svc.Delete(c.Request().Context(), id); err != nil {
		if handled := writeServiceError(c, err); handled {
			return nil
		}
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": err.Error()})
	}
	return c.NoContent(http.StatusNoContent)
}

type createUserRequest struct {
	Name  string `json:"name"`
	Email string `json:"email"`
}

type updateUserRequest struct {
	Name  string `json:"name"`
	Email string `json:"email"`
}

type userDTO struct {
	ID        int64  `json:"id"`
	Name      string `json:"name"`
	Email     string `json:"email"`
	Score     int64  `json:"score"`
	CreatedAt string `json:"created_at"`
	UpdatedAt string `json:"updated_at"`
}

const timeRFC3339 = "2006-01-02T15:04:05Z07:00"

func toDTO(u *models.User) userDTO {
	return userDTO{
		ID:        u.ID,
		Name:      u.Name,
		Email:     u.Email,
		Score:     u.Score,
		CreatedAt: u.CreatedAt.UTC().Format(timeRFC3339),
		UpdatedAt: u.UpdatedAt.UTC().Format(timeRFC3339),
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
