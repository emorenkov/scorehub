package rest

import (
	"context"
	"errors"
	"net/http"

	"github.com/emorenkov/scorehub/pkg/common/middleware"
	"github.com/emorenkov/scorehub/pkg/common/models"
	"github.com/emorenkov/scorehub/pkg/user/config"
	"github.com/emorenkov/scorehub/pkg/user/service"
	"github.com/labstack/echo/v4"
	echoMiddleware "github.com/labstack/echo/v4/middleware"
	"github.com/redis/go-redis/v9"
	"go.uber.org/zap"
)

type Server struct {
	cfg     *config.UserConfig
	svc     service.User
	log     *zap.Logger
	e       *echo.Echo
	redis   *redis.Client
	limiter *middleware.RateLimiter
}

func NewServer(cfg *config.UserConfig, svc service.User, log *zap.Logger) (*Server, error) {
	if cfg.APIKey == "" {
		return nil, errors.New("API_KEY must be set")
	}

	e := echo.New()
	e.HideBanner = true
	e.HidePort = true

	redisClient := newRedisClient(cfg.RedisConfig)
	limiter := middleware.NewRateLimiter(redisClient, cfg.RateLimitBurst)

	s := &Server{
		cfg:     cfg,
		svc:     svc,
		log:     log,
		e:       e,
		redis:   redisClient,
		limiter: limiter,
	}

	e.Use(echoMiddleware.Recover())
	if limiter != nil {
		e.Use(middleware.EchoRateLimit(limiter))
	}

	s.registerRoutes()
	return s, nil
}

func (s *Server) registerRoutes() {
	s.e.GET("/_health", func(c echo.Context) error {
		return c.NoContent(http.StatusOK)
	})

	api := s.e.Group("/api/v1")
	api.POST("/users", s.createUser)

	protected := api.Group("", s.keyAuthMiddleware)
	protected.GET("/users/:id", s.getUser)
	protected.GET("/users", s.listUsers)
	protected.PUT("/users/:id", s.updateUser)
	protected.DELETE("/users/:id", s.deleteUser)
}

func (s *Server) Serve() error {
	addr := ":" + s.cfg.HTTPPort
	s.log.Info("starting REST server", zap.String("addr", addr))
	return s.e.Start(addr)
}

func (s *Server) Shutdown(ctx context.Context) error {
	s.log.Info("shutting down REST server")
	if s.redis != nil {
		_ = s.redis.Close()
	}
	return s.e.Shutdown(ctx)
}

func (s *Server) keyAuthMiddleware(next echo.HandlerFunc) echo.HandlerFunc {
	return func(c echo.Context) error {
		apiKey := c.Request().Header.Get("X-API-Key")
		if apiKey == "" || apiKey != s.cfg.APIKey {
			return c.NoContent(http.StatusUnauthorized)
		}
		return next(c)
	}
}

func newRedisClient(cfg *models.RedisConfig) *redis.Client {
	if cfg.RedisAddr == "" {
		return nil
	}
	return redis.NewClient(&redis.Options{
		Addr:     cfg.RedisAddr,
		Password: cfg.RedisPassword,
		DB:       cfg.RedisDB,
	})
}
