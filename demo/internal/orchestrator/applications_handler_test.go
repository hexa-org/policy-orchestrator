package orchestrator_test

import (
	"bytes"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"net"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"testing"

	"github.com/hexa-org/policy-mapper/api/policyprovider"
	"github.com/hexa-org/policy-mapper/pkg/healthsupport"
	"github.com/hexa-org/policy-mapper/pkg/hexapolicy"
	"github.com/hexa-org/policy-mapper/pkg/mockOidcSupport"
	"github.com/hexa-org/policy-mapper/pkg/oauth2support"
	"github.com/hexa-org/policy-mapper/pkg/websupport"
	"github.com/hexa-org/policy-mapper/sdk"
	"github.com/hexa-org/policy-orchestrator/demo/internal/orchestrator"
	"github.com/hexa-org/policy-orchestrator/demo/internal/orchestrator/test"
	"github.com/hexa-org/policy-orchestrator/demo/pkg/dataConfigGateway"
	"github.com/hexa-org/policy-orchestrator/demo/pkg/testsupport"
	"github.com/stretchr/testify/assert"
)

type applicationsHandlerData struct {
	testDir           string
	Data              *dataConfigGateway.ConfigData
	gateway           dataConfigGateway.ApplicationsDataGateway
	server            *http.Server
	key               string
	providers         map[string]policyprovider.Provider
	applicationTestId string
	MockOauth         *mockOidcSupport.MockAuthServer
	oauthHttpClient   *http.Client
}

func (data *applicationsHandlerData) SetUp() {
	_ = os.Setenv(sdk.EnvTestProvider, sdk.ProviderTypeMock)
	tempDir, _ := os.MkdirTemp("", "hexa-orchestrator-*")
	data.testDir = tempDir
	// The Mock Authorization Server is needed to issue tokens, and provide a JWKS endpoint for validation
	data.MockOauth = mockOidcSupport.NewMockAuthServer("clientId", "secret", map[string]interface{}{})

	mockUrlJwks, _ := url.JoinPath(data.MockOauth.Server.URL, "/jwks")
	tokenUrl, _ := url.JoinPath(data.MockOauth.Server.URL, "/token")
	// Set Env for Jwt Token Validation by Orchestrator handlers
	_ = os.Setenv(oauth2support.EnvOAuthJwksUrl, mockUrlJwks)
	_ = os.Setenv(oauth2support.EnvJwtRealm, "TEST_REALM")
	_ = os.Setenv(oauth2support.EnvJwtAuth, "true")

	// Set Env for Orchestrator Client
	_ = os.Setenv(oauth2support.EnvOAuthClientId, "clientId")
	_ = os.Setenv(oauth2support.EnvOAuthClientSecret, "secret")
	_ = os.Setenv(oauth2support.EnvOAuthTokenEndpoint, tokenUrl)

	jwtClientHandler := oauth2support.NewJwtClientHandler()
	data.oauthHttpClient = jwtClientHandler.GetHttpClient()

	testConfigPath := filepath.Join(data.testDir, ".hexa", "config.json")

	_ = os.Setenv(dataConfigGateway.EnvIntegrationConfigFile, testConfigPath)

	data.Data, _ = dataConfigGateway.NewIntegrationConfigData()
	data.gateway = data.Data.GetApplicationDataGateway()

	_, err := data.Data.Create("aName", "noop", []byte("aKey"))
	if err != nil {
		panic(err)
	}
	_, err = data.Data.Create("yetAnotherName", "zone_cloud", []byte("aKey"))
	if err != nil {
		panic(err)
	}

	listener, _ := net.Listen("tcp", "localhost:0")
	addr := listener.Addr().String()

	hash := sha256.Sum256([]byte("aKey"))
	data.key = hex.EncodeToString(hash[:])

	data.providers = make(map[string]policyprovider.Provider)
	data.providers["yetAnotherName"] = &orchestratorNoopProvider.NoopProvider{}
	data.providers["aName"] = &orchestratorNoopProvider.NoopProvider{}

	handlers := orchestrator.LoadHandlers(data.Data, data.providers)
	data.server = websupport.Create(addr, handlers, websupport.Options{})
	go websupport.Start(data.server, listener)
	healthsupport.WaitForHealthy(data.server)
	apps, _ := data.gateway.Find(true)
	data.applicationTestId = apps[0].ID
}

func (data *applicationsHandlerData) TearDown() {
	data.oauthHttpClient.CloseIdleConnections()
	data.MockOauth.Shutdown()
	websupport.Stop(data.server)

	data.key = ""
	data.Data = nil
	data.providers = nil
	data.server = nil
	data.gateway = nil
	data.oauthHttpClient = nil
	data.MockOauth = nil

	_ = os.Unsetenv(oauth2support.EnvOAuthTokenEndpoint)
	_ = os.Unsetenv(oauth2support.EnvOAuthJwksUrl)

	_ = os.RemoveAll(data.testDir)
	data.testDir = ""
}

func TestListApps(t *testing.T) {
	testsupport.WithSetUp(&applicationsHandlerData{}, func(data *applicationsHandlerData) {
		reqUrl := fmt.Sprintf("http://%s/applications", data.server.Addr)

		resp, _ := data.oauthHttpClient.Get(reqUrl)
		assert.Equal(t, http.StatusOK, resp.StatusCode)

		var apps orchestrator.Applications
		_ = json.NewDecoder(resp.Body).Decode(&apps)
		assert.Equal(t, 2, len(apps.Applications))

		application := apps.Applications[0]
		assert.Equal(t, "somewhereToMock", application.ObjectId)
		assert.Equal(t, "mock", application.Name)
		assert.Equal(t, "noop", application.ProviderName)
		assert.Equal(t, "Mock PAP", application.Description)
		assert.Equal(t, "noop", application.Service)
	})
}

func TestListApps_withSort(t *testing.T) {
	testsupport.WithSetUp(&applicationsHandlerData{}, func(data *applicationsHandlerData) {
		reqUrl := fmt.Sprintf("http://%s/applications", data.server.Addr)
		resp, _ := data.oauthHttpClient.Get(reqUrl)
		assert.Equal(t, http.StatusOK, resp.StatusCode)

		var apps orchestrator.Applications
		_ = json.NewDecoder(resp.Body).Decode(&apps)
		assert.Equal(t, 2, len(apps.Applications))

		expApps := []orchestrator.Application{
			{ProviderName: "noop", Service: "noop"},
			{ProviderName: "zone_cloud", Service: "zone_cloud"}}

		for a, actApp := range apps.Applications {
			assert.Equal(t, expApps[a].ProviderName, actApp.ProviderName)
			assert.Equal(t, expApps[a].Service, actApp.Service)
		}
	})
}

func TestShowApps(t *testing.T) {
	testsupport.WithSetUp(&applicationsHandlerData{}, func(data *applicationsHandlerData) {
		reqUrl := fmt.Sprintf("http://%s/applications/%s", data.server.Addr, data.applicationTestId)

		resp, _ := data.oauthHttpClient.Get(reqUrl)
		assert.Equal(t, http.StatusOK, resp.StatusCode)

		var app orchestrator.Application
		_ = json.NewDecoder(resp.Body).Decode(&app)

		assert.Equal(t, "somewhereToMock", app.ObjectId)
		assert.Equal(t, "mock", app.Name)
		assert.Equal(t, "Mock PAP", app.Description)
		assert.Equal(t, "noop", app.Service)
	})
}

func TestShowApps_withUnknownID(t *testing.T) {
	testsupport.WithSetUp(&applicationsHandlerData{}, func(data *applicationsHandlerData) {
		reqUrl := fmt.Sprintf("http://%s/applications/oops", data.server.Addr)

		resp, _ := data.oauthHttpClient.Get(reqUrl)
		assert.Equal(t, http.StatusInternalServerError, resp.StatusCode)
	})
}

func TestGetPolicies(t *testing.T) {
	testsupport.WithSetUp(&applicationsHandlerData{}, func(data *applicationsHandlerData) {
		reqUrl := fmt.Sprintf("http://%s/applications/%s/policies", data.server.Addr, data.applicationTestId)

		resp, _ := data.oauthHttpClient.Get(reqUrl)
		assert.Equal(t, http.StatusOK, resp.StatusCode)

		var policies hexapolicy.Policies
		_ = json.NewDecoder(resp.Body).Decode(&policies)
		assert.Equal(t, 2, len(policies.Policies))

		policy := policies.Policies[0]
		assert.Equal(t, "anAction", policy.Actions[0].String())
		assert.Equal(t, "aVersion", policy.Meta.Version)
		assert.Equal(t, []string{"user:aUser"}, policy.Subjects.String())
		assert.Equal(t, "anId", policy.Object.String())
	})
}

func TestGetPolicies_withFailedRequest(t *testing.T) {
	testsupport.WithSetUp(&applicationsHandlerData{}, func(data *applicationsHandlerData) {

		provider := data.providers["aName"]
		noop := provider.(*orchestratorNoopProvider.NoopProvider)
		noop.SetTestErr(errors.New("oops"))
		provider = data.providers["yetAnotherName"]
		noop = provider.(*orchestratorNoopProvider.NoopProvider)
		noop.SetTestErr(errors.New("oops"))
		reqUrl := fmt.Sprintf("http://%s/applications/%s/policies", data.server.Addr, data.applicationTestId)

		resp, _ := data.oauthHttpClient.Get(reqUrl)
		assert.Equal(t, http.StatusInternalServerError, resp.StatusCode)

		// reset
		noop.SetTestErr(nil)
	})
}

func TestSetPolicies(t *testing.T) {
	testsupport.WithSetUp(&applicationsHandlerData{}, func(data *applicationsHandlerData) {
		var buf bytes.Buffer
		policy := hexapolicy.PolicyInfo{
			Meta:     hexapolicy.MetaInfo{Version: "0.7"},
			Actions:  []hexapolicy.ActionInfo{"anAction"},
			Subjects: []string{"anEmail", "anotherEmail"},
			Object:   "aResourceId",
		}
		_ = json.NewEncoder(&buf).Encode(hexapolicy.Policies{Policies: []hexapolicy.PolicyInfo{policy}})

		reqUrl := fmt.Sprintf("http://%s/applications/%s/policies", data.server.Addr, data.applicationTestId)

		resp, _ := data.oauthHttpClient.Post(reqUrl, "application/json", bytes.NewReader(buf.Bytes()))
		assert.Equal(t, http.StatusCreated, resp.StatusCode)
	})
}

func TestSetPolicies_withErroneousProvider(t *testing.T) {
	testsupport.WithSetUp(&applicationsHandlerData{}, func(data *applicationsHandlerData) {
		provider := data.providers["aName"]
		noop := provider.(*orchestratorNoopProvider.NoopProvider)
		noop.SetTestErr(errors.New("oops"))

		var buf bytes.Buffer
		policy := hexapolicy.PolicyInfo{Meta: hexapolicy.MetaInfo{Version: "0.7"}, Actions: []hexapolicy.ActionInfo{"anAction"}, Subjects: []string{"user:anEmail", "user:anotherEmail"}, Object: "aResourceId"}
		_ = json.NewEncoder(&buf).Encode(policy)

		reqUrl := fmt.Sprintf("http://%s/applications/%s/policies", data.server.Addr, data.applicationTestId)

		resp, _ := data.oauthHttpClient.Post(reqUrl, "application/json", bytes.NewReader(buf.Bytes()))
		assert.Equal(t, http.StatusInternalServerError, resp.StatusCode)

		// reset after test
		noop.SetTestErr(nil)
	})
}

func TestSetPolicies_withMissingJson(t *testing.T) {
	testsupport.WithSetUp(&applicationsHandlerData{}, func(data *applicationsHandlerData) {
		reqUrl := fmt.Sprintf("http://%s/applications/%s/policies", data.server.Addr, "anId")

		resp, _ := data.oauthHttpClient.Post(reqUrl, "application/json", nil)
		assert.Equal(t, http.StatusInternalServerError, resp.StatusCode)
	})
}
