package restful

import (
	"net/http"
	"net/url"
	"testing"

	"github.com/stretchr/testify/suite"
	"github.com/vx416/dcard-work/internal/testutil"
	apiv1 "github.com/vx416/dcard-work/pkg/api/v1"
	"github.com/vx416/dcard-work/pkg/factory"
)

func TestAnimalEndpoint(t *testing.T) {
	suite.Run(t, &animalEndpointSuite{new(handlerSuite)})
}

type animalEndpointSuite struct {
	*handlerSuite
}

func (s *animalEndpointSuite) assertGetAnimal(reqData *apiv1.GetGuardianAnimalRequest) *apiv1.GetGuardianAnimalResponse {
	query := url.QueryEscape(reqData.Name)
	req, record := testutil.BuildRequest("GET", "/guardian_animal?name="+query, nil)
	s.serv.ServeHTTP(record, req)
	s.Assert().Equal(http.StatusOK, record.Code)
	resp := &apiv1.GetGuardianAnimalResponse{}
	err := testutil.GetResponseData(record, resp)
	s.Require().NoError(err)
	s.Assert().NotEmpty(resp.Animal)
	s.Assert().NotEmpty(resp.Description)
	s.Assert().Equal(resp.Name, reqData.Name)
	return resp
}

func (s *animalEndpointSuite) Test_GetSameGuardianAnimal() {
	reqData := factory.GuardianAnimalRequest.MustBuild().(*apiv1.GetGuardianAnimalRequest)

	animal := ""
	description := ""
	for i := 0; i < 5; i++ {
		resp := s.assertGetAnimal(reqData)
		if animal == "" {
			animal = resp.Animal
			description = resp.Description
		} else {
			s.Assert().Equal(animal, resp.Animal)
			s.Assert().Equal(description, resp.Description)
		}
	}
}

func (s *animalEndpointSuite) Test_GetDiffAnimal() {
	reqData := factory.GuardianAnimalRequest.MustBuild().(*apiv1.GetGuardianAnimalRequest)
	compare := s.assertGetAnimal(reqData)
	diff := false
	for i := 0; i < 10; i++ {
		reqData := factory.GuardianAnimalRequest.MustBuild().(*apiv1.GetGuardianAnimalRequest)
		resp := s.assertGetAnimal(reqData)
		if resp.Animal != compare.Animal {
			diff = true
			break
		}
	}
	s.Assert().True(diff)
}
