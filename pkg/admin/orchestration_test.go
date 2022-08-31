package admin_test

import (
	"errors"
	"fmt"
	"github.com/hexa-org/policy-orchestrator/pkg/admin"
	"github.com/hexa-org/policy-orchestrator/pkg/admin/test"
	"github.com/hexa-org/policy-orchestrator/pkg/healthsupport"
	"github.com/hexa-org/policy-orchestrator/pkg/websupport"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/suite"
	"io"
	"net"
	"net/http"
	"testing"
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
	suite.client = new(admin_test.MockClient)
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
	resp, _ := http.Get(fmt.Sprintf("http://%s/orchestration/new", suite.server.Addr))
	body, _ := io.ReadAll(resp.Body)
	assert.Contains(suite.T(), string(body), "Policy Orchestration")
	assert.Contains(suite.T(), string(body), "Apply from")
	assert.Contains(suite.T(), string(body), "Apply to")
}

func (suite *OrchestrationSuite) TestNewOrchestration_withClientError() {
	suite.client.Errs = map[string]error{"http://noop/applications": errors.New("oops")}

	resp, _ := http.Get(fmt.Sprintf("http://%s/orchestration/new", suite.server.Addr))
	body, _ := io.ReadAll(resp.Body)
	assert.Contains(suite.T(), string(body), "Something went wrong.")
}

func (suite *OrchestrationSuite) TestCreateOrchestration() {
	resp, _ := http.Post(fmt.Sprintf("http://%s/orchestration", suite.server.Addr), "application/json", nil)
	_, _ = io.ReadAll(resp.Body)
	assert.Equal(suite.T(), http.StatusOK, resp.StatusCode)
}
