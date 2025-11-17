package rest

import (
	"encoding/json"
	"strconv"

	apperrors "github.com/emorenkov/scorehub/pkg/common/errors"
	"github.com/emorenkov/scorehub/pkg/common/models"
	"github.com/valyala/fasthttp"
)

func (s *Server) createUser(ctx *fasthttp.RequestCtx) {
	var req createUserRequest
	if err := json.Unmarshal(ctx.PostBody(), &req); err != nil {
		writeJSON(ctx, fasthttp.StatusBadRequest, map[string]string{"error": "invalid json"})
		return
	}
	u, err := s.svc.Create(ctx, req.Name, req.Email)
	if err != nil {
		if !writeServiceError(ctx, err) {
			writeJSON(ctx, fasthttp.StatusInternalServerError, map[string]string{"error": err.Error()})
		}
		return
	}
	writeJSON(ctx, fasthttp.StatusCreated, toDTO(u))
}

func (s *Server) getUser(ctx *fasthttp.RequestCtx) {
	id, ok := parseID(ctx)
	if !ok {
		writeJSON(ctx, fasthttp.StatusBadRequest, map[string]string{"error": "invalid id"})
		return
	}
	u, err := s.svc.Get(ctx, id)
	if err != nil {
		if !writeServiceError(ctx, err) {
			writeJSON(ctx, fasthttp.StatusInternalServerError, map[string]string{"error": err.Error()})
		}
		return
	}
	writeJSON(ctx, fasthttp.StatusOK, toDTO(u))
}

func (s *Server) listUsers(ctx *fasthttp.RequestCtx) {
	users, err := s.svc.List(ctx)
	if err != nil {
		if !writeServiceError(ctx, err) {
			writeJSON(ctx, fasthttp.StatusInternalServerError, map[string]string{"error": err.Error()})
		}
		return
	}
	resp := make([]userDTO, 0, len(users))
	for i := range users {
		resp = append(resp, toDTO(&users[i]))
	}
	writeJSON(ctx, fasthttp.StatusOK, resp)
}

func (s *Server) updateUser(ctx *fasthttp.RequestCtx) {
	id, ok := parseID(ctx)
	if !ok {
		writeJSON(ctx, fasthttp.StatusBadRequest, map[string]string{"error": "invalid id"})
		return
	}
	var req updateUserRequest
	if err := json.Unmarshal(ctx.PostBody(), &req); err != nil {
		writeJSON(ctx, fasthttp.StatusBadRequest, map[string]string{"error": "invalid json"})
		return
	}

	u, err := s.svc.Update(ctx, id, req.Name, req.Email)
	if err != nil {
		if !writeServiceError(ctx, err) {
			writeJSON(ctx, fasthttp.StatusInternalServerError, map[string]string{"error": err.Error()})
		}
		return
	}
	writeJSON(ctx, fasthttp.StatusOK, toDTO(u))
}

func (s *Server) deleteUser(ctx *fasthttp.RequestCtx) {
	id, ok := parseID(ctx)
	if !ok {
		writeJSON(ctx, fasthttp.StatusBadRequest, map[string]string{"error": "invalid id"})
		return
	}
	if err := s.svc.Delete(ctx, id); err != nil {
		if !writeServiceError(ctx, err) {
			writeJSON(ctx, fasthttp.StatusInternalServerError, map[string]string{"error": err.Error()})
		}
		return
	}
	ctx.SetStatusCode(fasthttp.StatusNoContent)
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

func parseID(ctx *fasthttp.RequestCtx) (int64, bool) {
	idStr := ctx.UserValue("id")
	if idStr == nil {
		return 0, false
	}
	id, err := strconv.ParseInt(idStr.(string), 10, 64)
	if err != nil {
		return 0, false
	}
	return id, true
}

func writeJSON(ctx *fasthttp.RequestCtx, status int, v any) {
	ctx.Response.Header.Set("Content-Type", "application/json")
	ctx.SetStatusCode(status)
	_ = json.NewEncoder(ctx).Encode(v)
}

func writeServiceError(ctx *fasthttp.RequestCtx, err error) bool {
	if err == nil {
		return false
	}
	if se, ok := apperrors.AsStatusError(err); ok {
		writeJSON(ctx, se.Status, map[string]string{"error": se.Message})
		return true
	}
	return false
}
