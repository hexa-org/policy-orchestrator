package orchestrator_test

import (
	"bytes"
	"encoding/json"
	"fmt"
	"github.com/hexa-org/policy-orchestrator/pkg/hawk_support"
	"github.com/hexa-org/policy-orchestrator/pkg/orchestrator"
	"github.com/hexa-org/policy-orchestrator/pkg/orchestrator/test"
	"github.com/hexa-org/policy-orchestrator/pkg/web_support"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/suite"
	"net/http"
	"testing"
)

type HandlerSuite struct {
	suite.Suite
	fields *orchestrator_test.SuiteFields
}

func TestIntegrationsHandler(t *testing.T) {
	suite.Run(t, &HandlerSuite{fields: &orchestrator_test.SuiteFields{}})
}

func (suite *HandlerSuite) SetupTest() {
	suite.fields.Setup()
	go web_support.Start(suite.fields.Server)
	web_support.WaitForHealthy(suite.fields.Server)
}

func (suite *HandlerSuite) TearDownTest() {
	_ = suite.fields.DB.Close()
	web_support.Stop(suite.fields.Server)
}

func (suite *HandlerSuite) TestList() {
	_, _ = suite.fields.Gateway.Create("aName", "google cloud", []byte("aKey"))
	resp, _ := hawk_support.HawkGet(&http.Client{}, "anId", suite.fields.Key, "http://localhost:8883/integrations")
	var jsonResponse orchestrator.Integrations
	_ = json.NewDecoder(resp.Body).Decode(&jsonResponse)

	assert.Equal(suite.T(), http.StatusOK, resp.StatusCode)
	assert.Equal(suite.T(), "aName", jsonResponse.Integrations[0].Name)
	assert.Equal(suite.T(), "google cloud", jsonResponse.Integrations[0].Provider)
	assert.Equal(suite.T(), []byte("aKey"), jsonResponse.Integrations[0].Key)
}

func (suite *HandlerSuite) TestCreate() {
	integration := orchestrator.Integration{Name: "aName", Provider: "google cloud", Key: []byte("aKey")}
	marshal, _ := json.Marshal(integration)
	_, _ = hawk_support.HawkPost(&http.Client{}, "anId", suite.fields.Key, "http://localhost:8883/integrations", bytes.NewReader(marshal))

	all, _ := suite.fields.Gateway.Find()
	assert.Equal(suite.T(), 1, len(all))
	assert.Equal(suite.T(), "aName", all[0].Name)
	assert.Equal(suite.T(), "google cloud", all[0].Provider)
	assert.Equal(suite.T(), []byte("aKey"), all[0].Key)
}

func (suite *HandlerSuite) TestDelete() {
	id, _ := suite.fields.Gateway.Create("aName", "google cloud", []byte("aKey"))
	resp, _ := hawk_support.HawkGet(&http.Client{}, "anId", suite.fields.Key, fmt.Sprintf("http://localhost:8883/integrations/%s", id))
	assert.Equal(suite.T(), resp.StatusCode, http.StatusOK)

	all, _ := suite.fields.Gateway.Find()
	assert.Equal(suite.T(), 0, len(all))
}
