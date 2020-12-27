package limiter

import (
	"context"
	"time"
)

type Pass struct {
	Allow         bool
	Err           error
	RemainingReqs float64
	TotalReqs     int64
	ResetAfter    time.Duration
}

type Limiter interface {
	Grant(ctx context.Context, key string) Pass
}

type Rate struct {
	period time.Duration
	reqs   float64

	reqPerSec float64
}

func (rate Rate) RefreshSec() float64 {
	return rate.period.Seconds()
}

func (r Rate) ReqPerSec() float64 {
	sec := r.period.Seconds()
	rate := r.reqs / float64(sec)

	return rate
}

func (r Rate) NumReqConsume(numReq float64) time.Duration {
	ratio := r.reqs / numReq
	sec := r.period.Seconds() / ratio
	sec = sec * 1e3
	return time.Duration(sec) * time.Millisecond
}

// Every
func Every(period time.Duration, reqs float64) Rate {
	return Rate{period: period, reqs: reqs}
}
