package middleware

import (
	"fmt"

	"github.com/labstack/echo/v4"
	"github.com/vx416/dcard-work/pkg/logging"
)

func Recovery(next echo.HandlerFunc) echo.HandlerFunc {
	return func(c echo.Context) error {
		defer func() {
			if r := recover(); r != nil {
				ctx := c.Request().Context()
				log := logging.Ctx(ctx)
				err, ok := r.(error)
				if !ok {
					err = fmt.Errorf("%v", r)
				}
				log.Err(err).Fatalf("[PANIC RECOVER] %+v", err)
				c.Error(err)
			}
		}()
		return next(c)
	}
}
