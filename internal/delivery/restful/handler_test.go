package restful

import (
	"context"
	"os"

	"github.com/go-redis/redis/v8"
	"github.com/stretchr/testify/suite"
	"github.com/vx416/dcard-work/internal/app"
	"github.com/vx416/dcard-work/internal/ring"
	"github.com/vx416/dcard-work/internal/testutil"
	"github.com/vx416/dcard-work/pkg/server"
	"go.uber.org/fx"
)

type handlerSuite struct {
	suite.Suite
	ti          *testutil.TestInstance
	serv        *server.Server
	redisClient *redis.Client
	ctx         context.Context
}

func (s *handlerSuite) SetupSuite() {
	os.Setenv("CONFIG_FILE", "app-test.yaml")

	ti, err := testutil.NewTestInstance()
	s.Require().NoError(err)
	s.ti = ti
	s.ctx = context.Background()

	options := fx.Options(
		fx.Provide(
			ring.DefaultConfig,
			ring.New,
			func(r *ring.HashRing) app.Ringer { return r },
			app.New,
			New,
		),
		fx.Invoke(V1Routes),
		fx.Populate(&s.redisClient),
	)

	serv, err := s.ti.NewServer(options)
	s.Require().NoError(err)
	s.serv = serv
}

func (s *handlerSuite) SetupTest() {
	err := s.redisClient.FlushDB(s.ctx).Err()
	s.Require().NoError(err)
}

func (s *handlerSuite) TearDownSuite() {
	err := s.ti.PruneAll()
	s.Require().NoError(err)
}
