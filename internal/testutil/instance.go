package testutil

import (
	"context"
	"errors"
	"path/filepath"
	"runtime"

	"github.com/go-redis/redis/v8"
	"github.com/vx416/dcard-work/pkg/config"
	"github.com/vx416/dcard-work/pkg/container"
	"github.com/vx416/dcard-work/pkg/limiter"
	"github.com/vx416/dcard-work/pkg/logging"
	"github.com/vx416/dcard-work/pkg/server"
	"go.uber.org/fx"
)

func getDataPath() string {
	_, f, _, _ := runtime.Caller(0)
	dirPath := filepath.Dir(f)
	return filepath.Join(dirPath, "../../configs/animals.json")
}

func NewTestInstance() (*TestInstance, error) {
	cfg, err := config.Init()
	if err != nil {
		return nil, err
	}
	builder, err := container.NewConBuilder()
	if err != nil {
		return nil, err
	}
	cfg.DataPath = getDataPath()

	return &TestInstance{
		cfg:     cfg,
		Builder: builder,
	}, nil
}

type TestInstance struct {
	*container.Builder
	cfg *config.Config
}

func (ti *TestInstance) runRedis() (*redis.Client, error) {
	return ti.RunRedis("ti-redis")
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
			ti.runRedis,
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
