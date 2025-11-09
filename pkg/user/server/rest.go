package server

import (
	"context"
	"encoding/json"
	"strconv"

	"github.com/emorenkov/scorehub/pkg/common/models"
	"github.com/emorenkov/scorehub/pkg/user/config"
	"github.com/fasthttp/router"
	"github.com/valyala/fasthttp"
	"go.uber.org/zap"
)

type Service interface {
	Create(ctx context.Context, name, email string) (*models.User, error)
	Get(ctx context.Context, id int64) (*models.User, error)
	List(ctx context.Context) ([]models.User, error)
	Update(ctx context.Context, id int64, name, email string) (*models.User, error)
	Delete(ctx context.Context, id int64) error
}

type Server struct {
	svc Service
	log *zap.Logger
	cfg *config.UserConfig
}

func NewServer(svc Service, log *zap.Logger, cfg *config.UserConfig) *Server {
	return &Server{
		svc: svc,
		log: log,
		cfg: cfg,
	}
}

// StartREST starts a fasthttp server with REST endpoints for the user service.
// It runs in a separate goroutine and stops when ctx is done.
func (s *Server) StartREST(ctxDone <-chan struct{}) {
	r := router.New()
	api := r.Group("/api/v1")
	api.POST("/users", s.createUser)
	api.GET("/users/{id}", s.getUser)
	api.GET("/users", s.listUsers)
	api.PUT("/users/{id}", s.updateUser)
	api.DELETE("/users/{id}", s.makeDeleteUser)

	srv := &fasthttp.Server{Handler: r.Handler}

	go func() {
		s.log.Info("starting REST server", zap.String("addr", ":"+s.cfg.HTTPPort))
		if err := srv.ListenAndServe(":" + s.cfg.HTTPPort); err != nil {
			s.log.Error("REST server exited", zap.Error(err))
		}
	}()

	go func() {
		<-ctxDone
		_ = srv.Shutdown()
		s.log.Info("REST server shutdown complete")
	}()
}

func writeJSON(ctx *fasthttp.RequestCtx, status int, v any) {
	ctx.Response.Header.Set("Content-Type", "application/json")
	ctx.SetStatusCode(status)
	_ = json.NewEncoder(ctx).Encode(v)
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

type createUserReq struct {
	Name  string `json:"name"`
	Email string `json:"email"`
}

type updateUserReq struct {
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

const timeRFC3339 = "2006-01-02T15:04:05Z07:00"

func (s *Server) createUser(ctx *fasthttp.RequestCtx) {
	var req createUserReq
	if err := json.Unmarshal(ctx.PostBody(), &req); err != nil {
		writeJSON(ctx, fasthttp.StatusBadRequest, map[string]string{"error": "invalid json"})
		return
	}
	u, err := s.svc.Create(context.Background(), req.Name, req.Email)
	if err != nil {
		writeJSON(ctx, fasthttp.StatusBadRequest, map[string]string{"error": err.Error()})
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
	u, err := s.svc.Get(context.Background(), id)
	if err != nil {
		writeJSON(ctx, fasthttp.StatusNotFound, map[string]string{"error": err.Error()})
		return
	}
	writeJSON(ctx, fasthttp.StatusOK, toDTO(u))
}

func (s *Server) listUsers(ctx *fasthttp.RequestCtx) {
	users, err := s.svc.List(context.Background())
	if err != nil {
		writeJSON(ctx, fasthttp.StatusInternalServerError, map[string]string{"error": err.Error()})
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
	var req updateUserReq
	if err := json.Unmarshal(ctx.PostBody(), &req); err != nil {
		writeJSON(ctx, fasthttp.StatusBadRequest, map[string]string{"error": "invalid json"})
		return
	}
	u, err := s.svc.Update(context.Background(), id, req.Name, req.Email)
	if err != nil {
		writeJSON(ctx, fasthttp.StatusBadRequest, map[string]string{"error": err.Error()})
		return
	}
	writeJSON(ctx, fasthttp.StatusOK, toDTO(u))
}

func (s *Server) makeDeleteUser(ctx *fasthttp.RequestCtx) {
	id, ok := parseID(ctx)
	if !ok {
		writeJSON(ctx, fasthttp.StatusBadRequest, map[string]string{"error": "invalid id"})
		return
	}
	if err := s.svc.Delete(context.Background(), id); err != nil {
		writeJSON(ctx, fasthttp.StatusInternalServerError, map[string]string{"error": err.Error()})
		return
	}
	ctx.SetStatusCode(fasthttp.StatusNoContent)
}
