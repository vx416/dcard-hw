package testutil

import (
	"context"
	"errors"

	"github.com/go-redis/redis/v8"
	"github.com/vx416/dcard-work/pkg/config"
	"github.com/vx416/dcard-work/pkg/container"
	"github.com/vx416/dcard-work/pkg/limiter"
	"github.com/vx416/dcard-work/pkg/logging"
	"github.com/vx416/dcard-work/pkg/server"
	"go.uber.org/fx"
)

func NewTestInstance() (*TestInstance, error) {
	cfg, err := config.Init()
	if err != nil {
		return nil, err
	}
	builder, err := container.NewConBuilder()
	if err != nil {
		return nil, err
	}

	return &TestInstance{
		cfg:     cfg,
		Builder: builder,
	}, nil
}

type TestInstance struct {
	*container.Builder
	cfg *config.Config
}

func (ti *TestInstance) NewServer(opt fx.Option) (*server.Server, error) {
	var (
		err  error
		serv = &server.Server{}
		ctx  = context.Background()
	)

	options := fx.Options(
		fx.Supply(*ti.cfg),
		fx.Provide(
			logging.New,
			ti.RunRedis,
			limiter.NewWithRedis,
			func(c *redis.Client) limiter.RedisScripter { return c },
			server.New,
		),
		fx.Populate(&serv),
	)

	if opt != nil {
		options = fx.Options(
			options,
			opt,
		)
	}

	app := fx.New(options)
	err = app.Start(ctx)
	if err != nil {
		return nil, err
	}
	if serv == nil {
		return nil, errors.New("initialize server failed")
	}
	err = app.Stop(ctx)
	if err != nil {
		return nil, err
	}

	return serv, nil
}
