package rest

import (
	"context"

	"github.com/emorenkov/scorehub/pkg/common/middleware"
	"github.com/emorenkov/scorehub/pkg/user/config"
	"github.com/emorenkov/scorehub/pkg/user/service"
	"github.com/fasthttp/router"
	"github.com/redis/go-redis/v9"
	"github.com/valyala/fasthttp"
	"go.uber.org/zap"
)

type Server struct {
	cfg        *config.UserConfig
	svc        service.Service
	log        *zap.Logger
	router     *router.Router
	httpServer *fasthttp.Server
	limiter    *middleware.RateLimiter
	redis      *redis.Client
}

func NewServer(cfg *config.UserConfig, svc service.Service, log *zap.Logger) *Server {
	r := router.New()
	s := &Server{
		cfg:     cfg,
		svc:     svc,
		log:     log,
		router:  r,
		redis:   newRedisClient(cfg),
		limiter: nil, // set below
		httpServer: &fasthttp.Server{
			Handler: nil, // set below after wrapping
		},
	}
	s.limiter = middleware.NewRateLimiter(s.redis, cfg.RateLimitRPS, cfg.RateLimitBurst)
	s.httpServer.Handler = s.wrap(r.Handler)
	s.registerRoutes()
	return s
}

func (s *Server) wrap(next fasthttp.RequestHandler) fasthttp.RequestHandler {
	return func(ctx *fasthttp.RequestCtx) {
		if !s.authenticate(ctx) {
			return
		}
		if !s.rateLimit(ctx) {
			return
		}
		next(ctx)
	}
}

func (s *Server) registerRoutes() {
	s.router.GET("/_health", func(ctx *fasthttp.RequestCtx) {
		ctx.SetStatusCode(fasthttp.StatusOK)
	})

	api := s.router.Group("/api/v1")
	api.POST("/users", s.createUser)
	api.GET("/users/{id}", s.getUser)
	api.GET("/users", s.listUsers)
	api.PUT("/users/{id}", s.updateUser)
	api.DELETE("/users/{id}", s.deleteUser)
}

func (s *Server) Serve() error {
	addr := ":" + s.cfg.HTTPPort
	s.log.Info("starting REST server", zap.String("addr", addr))
	return s.httpServer.ListenAndServe(addr)
}

func (s *Server) Shutdown(ctx context.Context) error {
	s.log.Info("shutting down REST server")
	if s.redis != nil {
		_ = s.redis.Close()
	}
	return s.httpServer.ShutdownWithContext(ctx)
}

func (s *Server) authenticate(ctx *fasthttp.RequestCtx) bool {
	if s.cfg.APIKey == "" {
		return true
	}
	if string(ctx.Request.Header.Peek("X-API-Key")) == s.cfg.APIKey {
		return true
	}
	ctx.SetStatusCode(fasthttp.StatusUnauthorized)
	return false
}

func (s *Server) rateLimit(ctx *fasthttp.RequestCtx) bool {
	if s.limiter.Allow(ctx.RemoteIP().String()) {
		return true
	}
	ctx.SetStatusCode(fasthttp.StatusTooManyRequests)
	return false
}

func newRedisClient(cfg *config.UserConfig) *redis.Client {
	if cfg.RedisAddr == "" {
		return nil
	}
	return redis.NewClient(&redis.Options{
		Addr:     cfg.RedisAddr,
		Password: cfg.RedisPassword,
		DB:       cfg.RedisDB,
	})
}
