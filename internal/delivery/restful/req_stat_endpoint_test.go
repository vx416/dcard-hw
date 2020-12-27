package restful

import (
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/stretchr/testify/suite"
	"github.com/vx416/dcard-work/internal/testutil"
	apiv1 "github.com/vx416/dcard-work/pkg/api/v1"
	"github.com/vx416/dcard-work/pkg/config"
)

func TestReqStatEndpoint(t *testing.T) {
	suite.Run(t, &reqEndpointSuite{new(handlerSuite)})
}

type reqEndpointSuite struct {
	*handlerSuite
}

func (s *reqEndpointSuite) assertPass(req *http.Request, record *httptest.ResponseRecorder) {
	s.serv.ServeHTTP(record, req)
	resp := &apiv1.ReqStatResponse{}
	err := testutil.GetResponseData(record, resp)
	s.Require().NoError(err)

	s.Assert().Equal(req.Header.Get("X-Real-IP"), resp.IP)
	s.Assert().Equal(record.Code, http.StatusOK)
	s.Assert().GreaterOrEqual(resp.RemainingRequest, int64(0))
}

func (s *reqEndpointSuite) assertFirstRequest(req *http.Request, record *httptest.ResponseRecorder) {
	s.serv.ServeHTTP(record, req)
	resp := &apiv1.ReqStatResponse{}
	err := testutil.GetResponseData(record, resp)
	s.Require().NoError(err)

	now := time.Now().Add(60 * time.Second).Unix()
	s.Assert().Equal(req.Header.Get("X-Real-IP"), resp.IP)
	s.Assert().Equal(int64(1), resp.RequestCount)
	s.Assert().Equal(int64(59), resp.RemainingRequest)
	s.Assert().Equal((60 * time.Second).String(), resp.ResetAfter)
	s.Assert().Equal(now, resp.ResetAt)
	s.Assert().Equal(record.Code, http.StatusOK)
}

func (s *reqEndpointSuite) assertBlockRequest(req *http.Request, record *httptest.ResponseRecorder, firstReqAt time.Time) {
	s.serv.ServeHTTP(record, req)
	s.Require().Equal(record.Code, http.StatusTooManyRequests)
	errResp := make(map[string]map[string]interface{})
	err := testutil.GetResponseData(record, &errResp)
	s.Require().NoError(err)
	s.Assert().NotEmpty(errResp["error"])
	s.Assert().Equal(errResp["error"]["code"], float64(http.StatusTooManyRequests))
	details := errResp["error"]["details"].(map[string]interface{})
	s.Assert().GreaterOrEqual(details["rateLimitRequestCount"], float64(61))
	s.Assert().Less(details["rateLimitRemainingRequest"], float64(1))
	s.Assert().Greater(details["rateLimitResetAt"], float64(firstReqAt.Unix()))
	s.T().Log(time.Unix(int64(details["rateLimitResetAt"].(float64)), 0))
}

func (s *reqEndpointSuite) Test_GetRequestStat() {
	req, record := testutil.BuildRequest("GET", "/", nil)
	s.assertFirstRequest(req, record)
}

func (s *reqEndpointSuite) Test_ReachLimit() {
	now := time.Now()
	for i := 0; i < 60; i++ {
		req, record := testutil.BuildRequest("GET", "/", nil)
		s.assertPass(req, record)
	}

	req, record := testutil.BuildRequest("GET", "/", nil)
	s.assertBlockRequest(req, record, now)
	req, record = testutil.BuildRequest("GET", "/", nil)
	req.Header.Set("X-Real-IP", "100.1.100.2")
	s.assertFirstRequest(req, record)

	time.Sleep(1 * time.Second)
	req, record = testutil.BuildRequest("GET", "/", nil)
	if config.Get().Limiter.Type == "counter" {
		s.assertBlockRequest(req, record, now)
	} else {
		s.assertPass(req, record)
	}

}
