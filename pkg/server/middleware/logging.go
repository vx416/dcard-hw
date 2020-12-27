package middleware

import (
	"time"

	"github.com/google/uuid"
	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
	"github.com/vx416/dcard-work/pkg/logging"
)

func reqIDGen() string {
	return uuid.New().String()
}

var reqIDConfig = middleware.RequestIDConfig{
	Generator: reqIDGen,
}

// ReqIDMiddleware request id middleware
var ReqIDMiddleware = middleware.RequestIDWithConfig(reqIDConfig)

func LoggingWithLogger(log logging.Logger) echo.MiddlewareFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			req := c.Request()
			ctx := req.Context()
			res := c.Response()
			rid := res.Header().Get(echo.HeaderXRequestID)

			c.SetRequest(req)

			start := time.Now()
			fields := map[string]interface{}{
				"request_id": rid,
				"method":     req.Method,
				"uri":        req.RequestURI,
				"bytes_in":   req.Header.Get(echo.HeaderContentLength),
				"query":      req.URL.RawQuery,
				"receive_at": start,
			}
			logger := log.With(fields)
			ctx = logger.Attach(ctx)
			req = req.WithContext(ctx)
			logger.Info("receive request")
			var err error
			if err = next(c); err != nil {
				c.Error(err)
			}

			stop := time.Now()
			logger = logger.Field("latency", stop.Sub(start).String()).Field("status", res.Status).Field("bytes_out", res.Size)
			if err != nil {
				logger.Err(err).Error("handle request failed")
			} else {
				logger.Info("handle request successfully")
			}
			return nil
		}
	}
}
