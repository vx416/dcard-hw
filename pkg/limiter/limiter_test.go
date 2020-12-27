package limiter

import (
	"context"
	"testing"
	"time"

	"github.com/go-redis/redis/v8"
	"github.com/stretchr/testify/assert"
	"github.com/vx416/dcard-work/pkg/container"
)

func newBuilder() (*container.Builder, error) {
	builder, err := container.NewConBuilder()
	if err != nil {
		return nil, err
	}
	return builder, nil
}

func newTestRedisClient(addr string, db int) (*redis.Client, error) {
	ctx := context.Background()
	opts := &redis.Options{}
	opts.Network = "tcp"
	opts.Addr = addr
	opts.DB = db
	client := redis.NewClient(opts)
	err := client.Ping(ctx).Err()
	if err != nil {
		return nil, err
	}
	err = client.FlushDB(ctx).Err()
	if err != nil {
		return nil, err
	}
	return client, nil
}

func testFirstPass(ctx context.Context, t *testing.T, limiter Limiter, tc testcase) {
	pass := limiter.Grant(ctx, "test")
	assert.NoError(t, pass.Err)
	assert.True(t, pass.Allow)
	assert.Equal(t, pass.TotalReqs, int64(1))
	assert.Equal(t, pass.ResetAfter, tc.rate.period)
	assert.Equal(t, pass.RemainingReqs, tc.burst-1)
}

func testReachLimit(ctx context.Context, t *testing.T, limiter Limiter, tc testcase) {
	for i := float64(0); i < tc.burst; i++ {
		pass := limiter.Grant(ctx, "test")
		assert.NoError(t, pass.Err)
		if pass.TotalReqs == int64(tc.burst) {
			assert.True(t, pass.Allow)
		}
	}
	pass := limiter.Grant(ctx, "test")
	assert.NoError(t, pass.Err)
	assert.False(t, pass.Allow)
	assert.GreaterOrEqual(t, pass.TotalReqs, int64(tc.burst))
	assert.Less(t, pass.RemainingReqs, float64(1))
}

func testWithOtherKey(ctx context.Context, t *testing.T, limiter Limiter, tc testcase) {
	pass := limiter.Grant(ctx, "test2")
	assert.NoError(t, pass.Err)
	assert.True(t, pass.Allow)
	assert.Equal(t, pass.TotalReqs, int64(1))
	assert.Equal(t, pass.ResetAfter, tc.rate.period)
	assert.Equal(t, pass.RemainingReqs, tc.burst-1)
}

func testWaitAMinute(ctx context.Context, t *testing.T, limiter Limiter, tc testcase) {
	wait := tc.rate.NumReqConsume(1)
	time.Sleep(wait)
	pass := limiter.Grant(ctx, "test")
	assert.NoError(t, pass.Err)

	switch tc.typ {
	case LeakyBucket:
		assert.True(t, pass.Allow)
	case SimpleCounter:
		assert.False(t, pass.Allow)
	}

	assert.Less(t, int64(pass.ResetAfter-wait), int64(tc.rate.period))
}

type testcase struct {
	name  string
	typ   LimiterType
	rate  Rate
	burst float64
}

func TestLimiter_Grant(t *testing.T) {
	builder, err := newBuilder()
	assert.NoError(t, err)
	client, err := builder.RunRedis("limiter-redis")
	assert.NoError(t, err)
	opts := client.Options()

	testcases := []testcase{
		{name: "leaky_bucket", typ: LeakyBucket, rate: Every(1*time.Minute, 60), burst: 60},
		{name: "counter", typ: SimpleCounter, rate: Every(30*time.Second, 60), burst: 60},
	}

	for _, tc := range testcases {
		t.Run(tc.name, func(t *testing.T) {
			ctx := context.Background()
			client, err := newTestRedisClient(opts.Addr, opts.DB+1)
			assert.NoError(t, err)
			limiter, err := newLimiter(tc.typ, tc.rate, tc.burst, client)
			assert.NoError(t, err)

			testFirstPass(ctx, t, limiter, tc)
			testReachLimit(ctx, t, limiter, tc)
			testWithOtherKey(ctx, t, limiter, tc)
			testWaitAMinute(ctx, t, limiter, tc)
		})
	}

	err = builder.PruneAll()
	assert.NoError(t, err)
}
