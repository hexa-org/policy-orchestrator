package admin_test

import (
	"errors"
	"fmt"
	"io"
	"net"
	"net/http"
	"testing"

	"github.com/hexa-org/policy-orchestrator/pkg/admin"
	"github.com/hexa-org/policy-orchestrator/pkg/admin/test"
	"github.com/hexa-org/policy-orchestrator/pkg/healthsupport"
	"github.com/hexa-org/policy-orchestrator/pkg/websupport"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/suite"
)

type OrchestrationSuite struct {
	suite.Suite
	server *http.Server
	client *admin_test.MockClient
}

func TestOrchestration(t *testing.T) {
	suite.Run(t, new(OrchestrationSuite))
}

func (suite *OrchestrationSuite) SetupTest() {
	listener, _ := net.Listen("tcp", "localhost:0")
	suite.client = &admin_test.MockClient{Url: "http://noop"}
	suite.server = websupport.Create(
		listener.Addr().String(),
		admin.LoadHandlers("http://noop", suite.client),
		websupport.Options{})
	go websupport.Start(suite.server, listener)
	healthsupport.WaitForHealthy(suite.server)
}

func (suite *OrchestrationSuite) TearDownTest() {
	websupport.Stop(suite.server)
}

func (suite *OrchestrationSuite) TestNewOrchestration() {
	suite.client.DesiredApplications = []admin.Application{
		{ID: "anId", IntegrationId: "anIntegrationId", ObjectId: "anObjectId", Name: "aName", Description: "aDescription", ProviderName: "google_cloud"},
		{ID: "anotherId", IntegrationId: "anotherIntegrationId", ObjectId: "anotherObjectId", Name: "anotherName", Description: "anotherDescription", ProviderName: "google_cloud"},
	}
	resp, _ := http.Get(fmt.Sprintf("http://%s/orchestration/new", suite.server.Addr))
	body, _ := io.ReadAll(resp.Body)
	assert.Contains(suite.T(), string(body), "Policy Orchestration")
	assert.Contains(suite.T(), string(body), "Apply from")
	assert.Contains(suite.T(), string(body), "<option value=\"anId\">")
	assert.Contains(suite.T(), string(body), "Apply to")
	assert.Contains(suite.T(), string(body), "<option value=\"anotherId\">")
}

func (suite *OrchestrationSuite) TestNewOrchestration_withClientError() {
	suite.client.Errs = map[string]error{"http://noop/applications": errors.New("oops")}

	resp, _ := http.Get(fmt.Sprintf("http://%s/orchestration/new", suite.server.Addr))
	body, _ := io.ReadAll(resp.Body)
	assert.Contains(suite.T(), string(body), "Something went wrong.")
}

func (suite *OrchestrationSuite) TestUpdateOrchestration() {
	resp, _ := http.Post(fmt.Sprintf("http://%s/orchestration", suite.server.Addr), "application/json", nil)
	_, _ = io.ReadAll(resp.Body)
	assert.Equal(suite.T(), http.StatusOK, resp.StatusCode)
}

func (suite *OrchestrationSuite) TestUpdateOrchestration_withError() {
	suite.client.Errs = map[string]error{"http://noop/orchestration": errors.New("oops")}

	resp, _ := http.Post(fmt.Sprintf("http://%s/orchestration", suite.server.Addr), "application/json", nil)
	body, _ := io.ReadAll(resp.Body)
	assert.Contains(suite.T(), string(body), "oops")
}
