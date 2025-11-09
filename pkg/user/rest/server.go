package rest

import (
	"context"

	"github.com/emorenkov/scorehub/pkg/user/config"
	"github.com/emorenkov/scorehub/pkg/user/service"
	"github.com/fasthttp/router"
	"github.com/valyala/fasthttp"
	"go.uber.org/zap"
)

type Server struct {
	cfg        *config.UserConfig
	svc        service.Service
	log        *zap.Logger
	router     *router.Router
	httpServer *fasthttp.Server
}

func NewServer(cfg *config.UserConfig, svc service.Service, log *zap.Logger) *Server {
	r := router.New()
	s := &Server{
		cfg:    cfg,
		svc:    svc,
		log:    log,
		router: r,
		httpServer: &fasthttp.Server{
			Handler: r.Handler,
		},
	}
	s.registerRoutes()
	return s
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
	return s.httpServer.ShutdownWithContext(ctx)
}
