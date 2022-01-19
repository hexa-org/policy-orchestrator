package orchestrator_test

import (
	"bytes"
	"encoding/json"
	"fmt"
	"github.com/hexa-org/policy-orchestrator/pkg/hawksupport"
	"github.com/hexa-org/policy-orchestrator/pkg/orchestrator"
	"github.com/hexa-org/policy-orchestrator/pkg/orchestrator/provider"
	"github.com/hexa-org/policy-orchestrator/pkg/orchestrator/test"
	"github.com/hexa-org/policy-orchestrator/pkg/websupport"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/suite"
	"net"
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
	listener, _ := net.Listen("tcp", "localhost:0")
	providers := make(map[string]provider.Provider)
	providers["google cloud"] = &orchestrator_test.NoopDiscovery{}
	suite.fields.Setup(providers, listener.Addr().String())
	go websupport.Start(suite.fields.Server, listener)
	websupport.WaitForHealthy(suite.fields.Server)
}

func (suite *HandlerSuite) TearDownTest() {
	_ = suite.fields.DB.Close()
	websupport.Stop(suite.fields.Server)
}

func (suite *HandlerSuite) TestList() {
	_, _ = suite.fields.Gateway.Create("aName", "google cloud", []byte("aKey"))
	resp, _ := hawksupport.HawkGet(&http.Client{}, "anId", suite.fields.Key, fmt.Sprintf("http://%s/integrations", suite.fields.Server.Addr))
	var jsonResponse orchestrator.Integrations
	_ = json.NewDecoder(resp.Body).Decode(&jsonResponse)

	assert.Equal(suite.T(), http.StatusOK, resp.StatusCode)
	assert.Equal(suite.T(), "aName", jsonResponse.Integrations[0].Name)
	assert.Equal(suite.T(), "google cloud", jsonResponse.Integrations[0].Provider)
	assert.Equal(suite.T(), []byte("aKey"), jsonResponse.Integrations[0].Key)
}

func (suite *HandlerSuite) TestCreate_fails() {
	integration := orchestrator.Integration{Name: "aName", Provider: "google cloud", Key: []byte("aKey")}
	marshal, _ := json.Marshal(integration)
	_, _ = hawksupport.HawkPost(&http.Client{}, "anId", suite.fields.Key, fmt.Sprintf("http://%s/integrations", suite.fields.Server.Addr), bytes.NewReader(marshal))

	all, _ := suite.fields.Gateway.Find()
	assert.Equal(suite.T(), 1, len(all))
	assert.Equal(suite.T(), "aName", all[0].Name)
	assert.Equal(suite.T(), "google cloud", all[0].Provider)
	assert.Equal(suite.T(), []byte("aKey"), all[0].Key)
}

func (suite *HandlerSuite) TestDelete() {
	id, _ := suite.fields.Gateway.Create("aName", "google cloud", []byte("aKey"))
	resp, _ := hawksupport.HawkGet(&http.Client{}, "anId", suite.fields.Key, fmt.Sprintf("http://%s/integrations/%s", suite.fields.Server.Addr, id))
	assert.Equal(suite.T(), resp.StatusCode, http.StatusOK)

	all, _ := suite.fields.Gateway.Find()
	assert.Equal(suite.T(), 0, len(all))
}

func (suite *HandlerSuite) TestDelete_bad_id() {
	_, _ = suite.fields.Gateway.Create("aName", "google cloud", []byte("aKey"))
	resp, _ := hawksupport.HawkGet(&http.Client{}, "anId", suite.fields.Key, fmt.Sprintf("http://%s/integrations/%s", suite.fields.Server.Addr, "0000"))
	assert.Equal(suite.T(), resp.StatusCode, http.StatusInternalServerError)
}
