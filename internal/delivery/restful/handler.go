package restful

import (
	"net/http"
	"time"

	"github.com/vx416/dcard-work/pkg/api/errors"
	"github.com/vx416/dcard-work/pkg/server/middleware"

	"github.com/labstack/echo/v4"
	"github.com/vx416/dcard-work/internal/app"
	apiv1 "github.com/vx416/dcard-work/pkg/api/v1"
)

func New(svc app.Servicer) *Handler {
	return &Handler{svc: svc}
}

type Handler struct {
	svc app.Servicer
}

func (h *Handler) ReqStatsEndpoint(c echo.Context) error {
	var (
		resp       = &apiv1.ReqStatResponse{}
		err        error
		respHeader = c.Response().Header()
	)

	resp.IP = c.RealIP()
	info, err := middleware.ParseRateLimitInfo(respHeader)
	if err != nil {
		return errors.Wrapf(errors.ErrInternalServerError, "handler: parse rate limit info failed, err:%+v", err)
	}
	resp.RequestCount = info[middleware.RateLimitRequest]
	resp.RemainingRequest = info[middleware.RateLimitRemaining]
	resp.ResetAfter = time.Duration(int(info[middleware.RateLimitResetAfter])).String()
	resp.ResetAt = info[middleware.RateLimitResetAt]

	return c.JSON(http.StatusOK, resp)
}

func (h *Handler) GetGuardianAnimalEndpoint(c echo.Context) error {
	var (
		ctx  = c.Request().Context()
		req  = &apiv1.GetGuardianAnimalRequest{}
		resp = &apiv1.GetGuardianAnimalResponse{}
		err  error
	)

	if err = c.Bind(req); err != nil {
		return errors.Wrapf(errors.ErrUnprocessableEntityError, "handler: parse request to struct failed, err:%+v", err)
	}

	if req.Name == "" {
		return errors.WithNewMsgf(errors.ErrUnprocessableEntityError, "name cannot be empty")
	}

	animal, err := h.svc.GetAnimal(ctx, req.Name)
	if err != nil {
		return err
	}
	resp.Animal = animal.Name
	resp.Description = animal.Description
	resp.Name = req.Name
	return c.JSON(http.StatusOK, resp)
}
