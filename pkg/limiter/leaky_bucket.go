package limiter

import (
	"context"
	"strconv"
	"time"

	"github.com/go-redis/redis/v8"
)

var leakeyBucket = redis.NewScript(`
local bucket_key = KEYS[1] .. ".bucket"
local total_req_key = KEYS[1] .. ".request_count"
local rate = tonumber(ARGV[1])
local bucket_size = tonumber(ARGV[2])
local request_counts = tonumber(ARGV[3])
local current_timestamp = tonumber(ARGV[4])
local refresh_interval = tonumber(ARGV[5])
local bucket_infos = redis.call("HGETALL", bucket_key)
local total_reqs = tonumber(redis.call("GET", total_req_key))
local current_size = tonumber(0)
local last_requested_at = current_timestamp
local set_expired = true

if next(bucket_infos) ~= nil then
	current_size = tonumber(bucket_infos[2])
	last_requested_at = tonumber(bucket_infos[4])
	set_expired = false
end

local elapsed = (current_timestamp - last_requested_at)
if elapsed < 0 then
	elapsed = 0
end

current_size = math.max(0, current_size-elapsed*rate)
local allowed = 2

if current_size+request_counts <= bucket_size then
	current_size = current_size + request_counts
	allowed = 1
	redis.call("HMSET", bucket_key, "current_size", current_size, "last_requested_at", current_timestamp)
end

if not total_reqs then
	total_reqs = 1
	redis.call("SETEX", total_req_key, refresh_interval, 1)
else
	total_reqs = total_reqs + 1
	redis.call("INCR", total_req_key)
end

if set_expired then
	redis.call("EXPIRE", bucket_key, refresh_interval)
end

local remaining_req = bucket_size - current_size


local ttl = redis.call("ttl", bucket_key)

return { allowed, tostring(remaining_req), total_reqs, ttl }
`)

func NewLeakBucket(rate Rate, maxSize float64, scripter RedisScripter) (Limiter, error) {

	err := leakeyBucket.Load(context.Background(), scripter).Err()
	if err != nil {
		return nil, err
	}

	return &LeakeyBucket{
		rate:     rate,
		maxSize:  maxSize,
		scripter: scripter,
		script:   leakeyBucket,
	}, nil
}

// LeakeyBucket token bucket rate limiter
type LeakeyBucket struct {
	maxSize float64
	rate    Rate

	scripter RedisScripter
	script   *redis.Script
}

func (limiter *LeakeyBucket) Grant(ctx context.Context, key string) Pass {
	var (
		pass = Pass{}
		keys = []string{key}
	)

	rate := limiter.rate.ReqPerSec()
	refreshSec := limiter.rate.RefreshSec()

	args := []interface{}{rate, limiter.maxSize, 1, time.Now().Unix(), refreshSec}

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
