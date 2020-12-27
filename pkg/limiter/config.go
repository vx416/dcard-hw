package limiter

import (
	"context"
	"fmt"
	"time"

	"github.com/go-redis/redis/v8"
)

type LimiterType string

const (
	LeakyBucket   LimiterType = "leaky_bucket"
	SimpleCounter LimiterType = "counter"
)

type Config struct {
	Type         LimiterType `yaml:"type" env:"TYPE"`
	Burst        float64     `yaml:"burst" env:"BURST"`
	PeriodSec    int64       `yaml:"period_sec" env:"PERIODSEC"`
	RequestCount float64     `yaml:"request_count" env:"REQUESTCOUNT"`
	Redis        struct {
		Host     string `yaml:"host" env:"REDIS_HOST"`
		Port     string `yaml:"port" env:"REDIS_PORT"`
		Password string `yaml:"password" env:"REDIS_PASSWORD"`
		DB       int    `yaml:"db" env:"REDIS_DB"`
	} `yaml:"redis"`
}

func NewWithRedis(cfg *Config, client RedisScripter) (Limiter, error) {
	period := time.Duration(cfg.PeriodSec) * time.Second

	burst := cfg.Burst
	if burst == 0 {
		burst = cfg.RequestCount
	}

	return newLimiter(cfg.Type, Every(period, cfg.RequestCount), burst, client)
}

func New(cfg *Config) (Limiter, error) {
	ctx := context.Background()
	client := newRedisClient(cfg)
	err := client.Ping(ctx).Err()
	if err != nil {
		return nil, err
	}

	period := time.Duration(cfg.PeriodSec) * time.Second

	burst := cfg.Burst
	if burst == 0 {
		burst = cfg.RequestCount
	}

	return newLimiter(cfg.Type, Every(period, cfg.RequestCount), burst, client)
}

func newLimiter(tye LimiterType, rate Rate, burst float64, client RedisScripter) (Limiter, error) {
	switch tye {
	case LeakyBucket:
		return NewLeakBucket(rate, burst, client)
	case SimpleCounter:
		return NewCounter(rate, burst, client)
	default:
		return nil, fmt.Errorf("limiter type(%s) not found", tye)
	}
}

func newRedisClient(cfg *Config) *redis.Client {
	opts := &redis.Options{}
	opts.Network = "tcp"
	opts.Addr = cfg.Redis.Host + ":" + cfg.Redis.Port
	opts.Password = cfg.Redis.Password
	opts.DB = cfg.Redis.DB
	return redis.NewClient(opts)
}
