package limiter

import (
	"context"
	"strconv"
	"time"

	"github.com/go-redis/redis/v8"
)

var simpleCounter = redis.NewScript(`
local counter_key = KEYS[1] .. ".counter"
local total_req_key = KEYS[1] .. ".request_count"
local max_request = tonumber(ARGV[1])
local request_count = tonumber(ARGV[2])
local refresh_interval = tonumber(ARGV[3])
local counter = redis.call("GET", counter_key)
local total_reqs = tonumber(redis.call("GET", total_req_key))
local set_expired = false

if not counter then
	counter = 0
	set_expired = true
end

local allowed = 2
if counter+request_count <= max_request then
	counter = counter + request_count
	allowed = 1
	redis.call("INCRBY", counter_key, request_count)
end

if not total_reqs then
	total_reqs = 1
	redis.call("SETEX", total_req_key, refresh_interval, 1)
else
	total_reqs = total_reqs + 1
	redis.call("INCR", total_req_key)
end

if set_expired then
	redis.call("EXPIRE", counter_key, refresh_interval)
end

local ttl = redis.call("ttl", counter_key)
local remaining_req = max_request - counter

return { allowed, tostring(remaining_req), total_reqs, ttl }
`)

// NewCounter construct counter limiter
func NewCounter(rate Rate, maxSize float64, scripter RedisScripter) (Limiter, error) {

	err := simpleCounter.Load(context.Background(), scripter).Err()
	if err != nil {
		return nil, err
	}

	return &Counter{
		rate:     rate,
		maxSize:  maxSize,
		scripter: scripter,
		script:   simpleCounter,
	}, nil
}

// Counter ...
type Counter struct {
	maxSize float64
	rate    Rate

	scripter RedisScripter
	script   *redis.Script
}

func (limiter *Counter) Grant(ctx context.Context, key string) Pass {
	var (
		pass = Pass{}
		keys = []string{key}
	)

	refreshSec := limiter.rate.RefreshSec()

	args := []interface{}{limiter.maxSize, 1, refreshSec}

	result, err := limiter.script.EvalSha(ctx, limiter.scripter, keys, args...).Result()
	if err != nil {
		pass.Err = err
		return pass
	}

	values := result.([]interface{})

	allowed := values[0].(int64)
	remaining := values[1].(string)
	totalReqs := values[2].(int64)
	reset := values[3].(int64)
	pass.Allow = allowed == 1
	pass.RemainingReqs, _ = strconv.ParseFloat(remaining, 64)
	pass.TotalReqs = totalReqs
	pass.ResetAfter = time.Duration(time.Duration(reset) * time.Second)
	return pass
}
