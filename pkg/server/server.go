package server

import (
	"context"
	"net/http"

	"github.com/vx416/dcard-work/pkg/limiter"

	"github.com/labstack/echo/v4"

	"github.com/vx416/dcard-work/pkg/logging"
	"github.com/vx416/dcard-work/pkg/server/middleware"
	"go.uber.org/fx"
)

type Config struct {
	Addr        string `yaml:"addr" env:"ADDR"`
	Port        string `yaml:"port" env:"PORT"`
	HealthCheck string `yaml:"healthcheck" env:"HEALTH_CHECK"`
	Mode        string `yaml:"mode" env:"MODE"`
}

type Server struct {
	*echo.Echo
	addr string
	port string
}

func (serv *Server) Addr() string {
	return serv.addr + ":" + serv.port
}

type Params struct {
	fx.In

	Config  *Config
	Log     logging.Logger
	Limiter limiter.Limiter
}

func RunWith(lc fx.Lifecycle, p Params) (*Server, error) {
	server, err := New(p)
	if err != nil {
		return nil, err
	}

	lc.Append(fx.Hook{
		OnStart: func(ctx context.Context) error {
			logging.Get().Debugf("server is listen on %s", server.Addr())
			go server.Start(server.Addr())
			return nil
		},
		OnStop: func(ctx context.Context) error {
			return server.Shutdown(ctx)
		},
	})

	return server, nil

}

func New(p Params) (*Server, error) {
	cfg := p.Config
	e := echo.New()

	if cfg.Mode == "debug" {
		e.Debug = true
		e.HideBanner = false
		e.HidePort = false
	} else {
		e.Debug = false
		e.HideBanner = true
		e.HidePort = true
	}

	e.Pre(middleware.ReqIDMiddleware)
	e.Use(
		middleware.LoggingWithLogger(p.Log),
		middleware.RateLimiterWithLimiter(p.Limiter),
		middleware.Recovery,
	)
	e.HTTPErrorHandler = EchoErrorHandler

	server := &Server{
		Echo: e,
		addr: cfg.Addr,
		port: cfg.Port,
	}

	healthCheck := "/ping"
	if cfg.HealthCheck != "" {
		healthCheck = cfg.HealthCheck
	}

	server.GET(healthCheck, func(c echo.Context) error {
		return c.String(http.StatusOK, "pong")
	})

	return server, nil
}
