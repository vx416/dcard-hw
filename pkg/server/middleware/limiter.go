package middleware

import (
	"net/http"
	"strconv"
	"time"

	"github.com/labstack/echo/v4"
	"github.com/vx416/dcard-work/pkg/api/errors"
	"github.com/vx416/dcard-work/pkg/limiter"
)

var (
	RateLimitRequest    = "X-RateLimit-Requests"
	RateLimitRemaining  = "X-RateLimit-Remaining"
	RateLimitResetAfter = "X-RateLimit-ResetAfter"
	RateLimitResetAt    = "X-RateLimit-ResetAt"
)

// RateLimiterWithLimiter rate limiter with limiter
func RateLimiterWithLimiter(lim limiter.Limiter) echo.MiddlewareFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			clientIP := c.RealIP()
			ctx := c.Request().Context()
			resp := c.Response()

			pass := lim.Grant(ctx, clientIP)
			if pass.Err != nil {
				return pass.Err
			}
			now := time.Now()
			resp.Header().Set(RateLimitRequest, strconv.Itoa(int(pass.TotalReqs)))
			resp.Header().Set(RateLimitRemaining, strconv.Itoa(int(pass.RemainingReqs)))
			resp.Header().Set(RateLimitResetAfter, strconv.Itoa(int(pass.ResetAfter)))
			resp.Header().Set(RateLimitResetAt, strconv.Itoa(int(now.Add(pass.ResetAfter).Unix())))
			if !pass.Allow {
				return errors.ErrTooManyRequestsError
			}
			return next(c)
		}
	}
}

func ParseRateLimitInfo(respHeader http.Header) (map[string]int64, error) {
	var (
		info = make(map[string]int64)
		err  error
	)
	info[RateLimitRequest], err = strconv.ParseInt(respHeader.Get("X-RateLimit-Requests"), 10, 64)
	if err != nil {
		return info, err
	}
	info[RateLimitRemaining], err = strconv.ParseInt(respHeader.Get("X-RateLimit-Remaining"), 10, 64)
	if err != nil {
		return info, err
	}
	info[RateLimitResetAfter], err = strconv.ParseInt(respHeader.Get("X-RateLimit-ResetAfter"), 10, 64)
	if err != nil {
		return info, err
	}
	info[RateLimitResetAt], err = strconv.ParseInt(respHeader.Get("X-RateLimit-ResetAt"), 10, 64)
	if err != nil {
		return info, err
	}
	return info, nil
}
