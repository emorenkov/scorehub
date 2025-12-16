package rest

import (
	"context"
	"net/http"

	"github.com/emorenkov/scorehub/pkg/event/config"
	"github.com/emorenkov/scorehub/pkg/event/service"
	"github.com/labstack/echo/v4"
	echoMiddleware "github.com/labstack/echo/v4/middleware"
	"go.uber.org/zap"
)

type Server struct {
	cfg *config.Config
	svc service.Event
	log *zap.Logger
	e   *echo.Echo
}

func NewServer(cfg *config.Config, svc service.Event, log *zap.Logger) *Server {
	e := echo.New()
	e.HideBanner = true
	e.HidePort = true

	s := &Server{
		cfg: cfg,
		svc: svc,
		log: log,
		e:   e,
	}

	e.Use(echoMiddleware.Recover())

	s.registerRoutes()
	return s
}

func (s *Server) registerRoutes() {
	s.e.GET("/_health", func(c echo.Context) error {
		return c.NoContent(http.StatusOK)
	})

	api := s.e.Group("/api/v1", s.keyAuthMiddleware)
	api.POST("/score-events", s.sendScoreEvent)
}

func (s *Server) Serve() error {
	addr := ":" + s.cfg.HTTPPort
	s.log.Info("starting REST server", zap.String("addr", addr))
	return s.e.Start(addr)
}

func (s *Server) Shutdown(ctx context.Context) error {
	s.log.Info("shutting down REST server")
	return s.e.Shutdown(ctx)
}

func (s *Server) keyAuthMiddleware(next echo.HandlerFunc) echo.HandlerFunc {
	if s.cfg.APIKey == "" {
		return next
	}
	return func(c echo.Context) error {
		if c.Request().Header.Get("X-API-Key") != s.cfg.APIKey {
			return c.NoContent(http.StatusUnauthorized)
		}
		return next(c)
	}
}
