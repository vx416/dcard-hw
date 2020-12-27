package server

import (
	"net/http"
	"time"

	"github.com/labstack/echo/v4"
	"github.com/vx416/dcard-work/pkg/api/errors"
	"github.com/vx416/dcard-work/pkg/logging"
	"github.com/vx416/dcard-work/pkg/server/middleware"
)

// EchoErrorHandler echo error handler
func EchoErrorHandler(err error, c echo.Context) {
	if err == nil {
		c.JSON(http.StatusOK, "")
		return
	}

	respErr := errors.ToErrResponse(err)
	respHeader := c.Response().Header()

	if respErr.HTTPStatus == http.StatusTooManyRequests {
		limitInfo, err := middleware.ParseRateLimitInfo(respHeader)
		if err != nil {
			logging.Ctx(c.Request().Context()).Fatalf("error handler handle response error failed, err:%+v", err)
		} else {
			respErr.AddDetail("rateLimitRequestCount", limitInfo[middleware.RateLimitRequest])
			respErr.AddDetail("rateLimitRemainingRequest", limitInfo[middleware.RateLimitRemaining])
			respErr.AddDetail("rateLimitRefreshAfter", time.Duration(limitInfo[middleware.RateLimitResetAfter]).String())
			respErr.AddDetail("rateLimitResetAt", limitInfo[middleware.RateLimitResetAt])
			respErr.AddDetail("rateLimitRequestIP", c.RealIP())
		}

	}

	reqID := respHeader.Get(echo.HeaderXRequestID)
	respErr.AddDetail("traceID", reqID)

	// Send response
	if !c.Response().Committed {
		if c.Request().Method == http.MethodHead { // Issue #608
			err = c.NoContent(respErr.HTTPStatus)
		} else {
			err = c.JSON(respErr.HTTPStatus, respErr)
		}
		if err != nil {
			logging.Ctx(c.Request().Context()).Fatalf("error handler handle response error failed, err:%+v", err)
		}
	}
}
