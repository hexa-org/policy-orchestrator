package orchestrator_test

import (
	"bytes"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"net"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"testing"

	"github.com/hexa-org/policy-mapper/api/policyprovider"
	"github.com/hexa-org/policy-mapper/sdk"
	"github.com/hexa-org/policy-orchestrator/demo/internal/orchestrator"
	orchestratortest "github.com/hexa-org/policy-orchestrator/demo/internal/orchestrator/test"
	"github.com/hexa-org/policy-orchestrator/demo/pkg/dataConfigGateway"
	"github.com/hexa-org/policy-orchestrator/demo/pkg/oauth2support"
	"github.com/hexa-org/policy-orchestrator/demo/pkg/oidctestsupport"

	"github.com/hexa-org/policy-orchestrator/demo/pkg/healthsupport"
	"github.com/hexa-org/policy-orchestrator/demo/pkg/websupport"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/suite"
)

type HandlerSuite struct {
	suite.Suite
	server          *http.Server
	key             string
	testDir         string
	Data            *dataConfigGateway.ConfigData
	gateway         dataConfigGateway.IntegrationsDataGateway
	MockOauth       *oidctestsupport.MockAuthServer
	oauthHttpClient *http.Client
}

func TestIntegrationsHandler(t *testing.T) {
	var err error
	s := HandlerSuite{}
	// The Mock Authorization Server is needed to issue tokens, and provide a JWKS endpoint for validation
	s.MockOauth = oidctestsupport.NewMockAuthServer("clientId", "secret", map[string]interface{}{})
	mockUrlJwks, _ := url.JoinPath(s.MockOauth.Server.URL, "/jwks")
	// Set Env for Jwt Token Validation by Orchestrator handlers
	_ = os.Setenv(oauth2support.EnvOAuthJwksUrl, mockUrlJwks)
	_ = os.Setenv(oauth2support.EnvJwtRealm, "TEST_REALM")
	_ = os.Setenv(oauth2support.EnvJwtAuth, "true")

	// Set Env for Orchestrator Client
	_ = os.Setenv(oauth2support.EnvOAuthClientId, "clientId")
	_ = os.Setenv(oauth2support.EnvOAuthClientSecret, "secret")
	_ = os.Setenv(oauth2support.EnvOAuthTokenEndpoint, fmt.Sprintf("%s/token", s.MockOauth.Server.URL))

	s.oauthHttpClient = oauth2support.NewJwtClientHandler().GetHttpClient()

	_ = os.Setenv(sdk.EnvTestProvider, sdk.ProviderTypeMock)
	dir, err := os.MkdirTemp("", "hexa-orchestrator-*")
	assert.NoError(t, err, "Error creating temp dir")

	s.testDir = dir

	testConfigPath := filepath.Join(s.testDir, ".hexa", "config.json")

	_ = os.Setenv(dataConfigGateway.EnvIntegrationConfigFile, testConfigPath)

	s.Data, err = dataConfigGateway.NewIntegrationConfigData()
	s.gateway = s.Data

	cache := make(map[string]policyprovider.Provider)
	cache["yetAnotherName"] = &orchestratortest.NoopProvider{}
	cache["aName"] = &orchestratortest.NoopProvider{}

	listener, _ := net.Listen("tcp", "localhost:0")
	addr := listener.Addr().String()

	hash := sha256.Sum256([]byte("aKey"))
	s.key = hex.EncodeToString(hash[:])

	handlers := orchestrator.LoadHandlers(s.Data, cache)
	s.server = websupport.Create(addr, handlers, websupport.Options{})

	go websupport.Start(s.server, listener)
	healthsupport.WaitForHealthy(s.server)

	if err == nil {
		suite.Run(t, &s)
	}

	s.oauthHttpClient.CloseIdleConnections()
	s.MockOauth.Shutdown()
	websupport.Stop(s.server)
	_ = os.RemoveAll(s.testDir)
	fmt.Println("** Test complete **")
}

func (s *HandlerSuite) TearDownTest() {
	s.Data.Integrations = make(map[string]*sdk.Integration)
}

func (s *HandlerSuite) TestList() {
	_, _ = s.gateway.Create("", "noop", []byte("aKey"))

	resp, _ := s.oauthHttpClient.Get(fmt.Sprintf("http://%s/integrations", s.server.Addr))
	assert.Equal(s.T(), http.StatusOK, resp.StatusCode)

	var jsonResponse orchestrator.Integrations
	_ = json.NewDecoder(resp.Body).Decode(&jsonResponse)

	integration := jsonResponse.Integrations[0]
	// assert.Equal(s.T(), "aName", integration.Name)
	assert.Equal(s.T(), "noop", integration.Provider)
	assert.Equal(s.T(), []byte("aKey"), integration.Key)
}

func (s *HandlerSuite) TestCreate() {
	integration := orchestrator.Integration{ID: "anId", Name: "aName", Provider: "noop", Key: []byte("aKey")}
	marshal, _ := json.Marshal(integration)
	_, _ = s.oauthHttpClient.Post(fmt.Sprintf("http://%s/integrations", s.server.Addr), "application/json", bytes.NewReader(marshal))

	records := s.gateway.Find()
	assert.Equal(s.T(), 1, len(records))

	record := records[0]
	assert.Equal(s.T(), "anId", record.ID)
	// assert.Equal(s.T(), "aName", record.Name)
	assert.Equal(s.T(), "noop", record.Provider)
	assert.Equal(s.T(), []byte("aKey"), record.Key)
}

func (s *HandlerSuite) TestDelete() {
	id, _ := s.gateway.Create("anId", "noop", []byte("aKey"))
	assert.Equal(s.T(), "anId", id)

	resp, _ := s.oauthHttpClient.Get(fmt.Sprintf("http://%s/integrations/%s", s.server.Addr, id))
	assert.Equal(s.T(), resp.StatusCode, http.StatusOK)

	records := s.gateway.Find()
	assert.Equal(s.T(), 0, len(records))
}

func (s *HandlerSuite) TestDelete_withUnknownID() {
	_, _ = s.gateway.Create("aName", "noop", []byte("aKey"))

	resp, _ := s.oauthHttpClient.Get(fmt.Sprintf("http://%s/integrations/%s", s.server.Addr, "0000"))
	assert.Equal(s.T(), resp.StatusCode, http.StatusInternalServerError)
}
