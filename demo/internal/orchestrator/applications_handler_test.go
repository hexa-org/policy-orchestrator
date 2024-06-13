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
	"sync"
	"testing"
	"time"

	"github.com/hexa-org/policy-mapper/api/policyprovider"
	"github.com/hexa-org/policy-mapper/pkg/hexapolicy"
	"github.com/hexa-org/policy-mapper/sdk"
	"github.com/hexa-org/policy-orchestrator/demo/internal/orchestrator"
	"github.com/hexa-org/policy-orchestrator/demo/internal/orchestrator/test"
	"github.com/hexa-org/policy-orchestrator/demo/pkg/dataConfigGateway"
	"github.com/hexa-org/policy-orchestrator/demo/pkg/healthsupport"
	"github.com/hexa-org/policy-orchestrator/demo/pkg/oauth2support"
	"github.com/hexa-org/policy-orchestrator/demo/pkg/oidctestsupport"
	"github.com/hexa-org/policy-orchestrator/demo/pkg/testsupport"
	"github.com/hexa-org/policy-orchestrator/demo/pkg/websupport"
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
	MockOauth         *oidctestsupport.MockAuthServer
	oauthHttpClient   *http.Client
	mu                sync.Mutex
}

func (data *applicationsHandlerData) SetUp() {
	data.mu.Lock()
	// The Mock Authorization Server is needed to issue tokens, and provide a JWKS endpoint for validation
	data.MockOauth = oidctestsupport.NewMockAuthServer("clientId", "secret", map[string]interface{}{})

	mockUrlJwks, _ := url.JoinPath(data.MockOauth.Server.URL, "/jwks")
	// Set Env for Jwt Token Validation by Orchestrator handlers
	_ = os.Setenv(oauth2support.EnvOAuthJwksUrl, mockUrlJwks)
	_ = os.Setenv(oauth2support.EnvJwtRealm, "TEST_REALM")
	_ = os.Setenv(oauth2support.EnvJwtAuth, "true")

	// Set Env for Orchestrator Client
	_ = os.Setenv(oauth2support.EnvOAuthClientId, "clientId")
	_ = os.Setenv(oauth2support.EnvOAuthClientSecret, "secret")
	_ = os.Setenv(oauth2support.EnvOAuthTokenEndpoint, fmt.Sprintf("%s/token", data.MockOauth.Server.URL))

	tempDir, _ := os.MkdirTemp("", "hexa-orchestrator-*")

	jwtClientHandler := oauth2support.NewJwtClientHandler()
	data.oauthHttpClient = jwtClientHandler.GetHttpClient()

	data.testDir = tempDir

	testConfigPath := filepath.Join(data.testDir, ".hexa", "config.json")
	_ = os.Setenv(sdk.EnvTestProvider, sdk.ProviderTypeMock)
	// _ = os.Unsetenv(sdk.EnvTestProvider)

	_ = os.Setenv(dataConfigGateway.EnvIntegrationConfigFile, testConfigPath)

	data.Data, _ = dataConfigGateway.NewIntegrationConfigData()
	data.gateway = data.Data.GetApplicationDataGateway()
	/*
	   	data.db, _ = databasesupport.Open("postgres://orchestrator:orchestrator@localhost:5432/orchestrator_test?sslmode=disable")
	   	_, _ = data.db.Exec(`
	   delete from applications;
	   delete from integrations;

	   insert into integrations (id, name, provider, key) values ('50e00619-9f15-4e85-a7e9-f26d87ea12e7', 'aName', 'noop', 'aKey');
	   insert into applications (id, integration_id, object_id, name, description, service) values ('6409776a-367a-483a-a194-5ccf9c4ff210', '50e00619-9f15-4e85-a7e9-f26d87ea12e7', 'anObjectId', 'aName', 'aDescription', 'aService');
	   insert into applications (id, integration_id, object_id, name, description, service) values ('6409776a-367a-483a-a194-5ccf9c4ff211', '50e00619-9f15-4e85-a7e9-f26d87ea12e7', 'anotherObjectId', 'anotherName', 'anotherDescription', 'anotherService');

	   insert into integrations (id, name, provider, key) values ('50e00619-9f15-4e85-a7e9-f26d87ea12e9', 'yetAnotherName', 'zone_cloud', 'aKey');
	   insert into applications (id, integration_id, object_id, name, description, service) values ('6409776a-367a-483a-a194-5ccf9c4ff212', '50e00619-9f15-4e85-a7e9-f26d87ea12e9', 'yetAnotherObjectId', 'yetAnotherName1', 'yetAnotherDescription', 'Kubernetes');
	   insert into applications (id, integration_id, object_id, name, description, service) values ('6409776a-367a-483a-a194-5ccf9c4ff213', '50e00619-9f15-4e85-a7e9-f26d87ea12e9', 'yetAnotherObjectId2', 'yetAnotherName2', 'yetAnotherDescription2', 'Container Kubernetes');
	   `)
	*/

	_, err := data.Data.Create("aName", "noop", []byte("aKey"))
	if err != nil {
		panic(err)
	}
	_, err = data.Data.Create("yetAnotherName", "zone_cloud", []byte("aKey"))

	listener, _ := net.Listen("tcp", "localhost:0")
	addr := listener.Addr().String()

	hash := sha256.Sum256([]byte("aKey"))
	data.key = hex.EncodeToString(hash[:])

	data.providers = make(map[string]policyprovider.Provider)
	data.providers["yetAnotherName"] = &orchestrator_test.NoopProvider{}
	data.providers["aName"] = &orchestrator_test.NoopProvider{}

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
	data.testDir = ""
	data.providers = nil
	data.server = nil
	data.gateway = nil
	data.mu.Unlock()
	_ = os.RemoveAll(data.testDir)
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
		assert.Equal(t, "anAction", policy.Actions[0].ActionUri)
		assert.Equal(t, "aVersion", policy.Meta.Version)
		assert.Equal(t, []string{"user:aUser"}, policy.Subject.Members)
		assert.Equal(t, "anId", policy.Object.ResourceID)
	})
}

func TestGetPolicies_withFailedRequest(t *testing.T) {
	testsupport.WithSetUp(&applicationsHandlerData{}, func(data *applicationsHandlerData) {

		provider := data.providers["aName"]
		noop := provider.(*orchestrator_test.NoopProvider)
		noop.SetTestErr(errors.New("oops"))
		provider = data.providers["yetAnotherName"]
		noop = provider.(*orchestrator_test.NoopProvider)
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
			Meta:    hexapolicy.MetaInfo{Version: "v0.5"},
			Actions: []hexapolicy.ActionInfo{{"anAction"}},
			Subject: hexapolicy.SubjectInfo{Members: []string{"anEmail", "anotherEmail"}},
			Object: hexapolicy.ObjectInfo{
				ResourceID: "aResourceId",
			},
		}
		_ = json.NewEncoder(&buf).Encode(hexapolicy.Policies{Policies: []hexapolicy.PolicyInfo{policy}})

		reqUrl := fmt.Sprintf("http://%s/applications/%s/policies", data.server.Addr, data.applicationTestId)

		resp, _ := data.oauthHttpClient.Post(reqUrl, "application/json", bytes.NewReader(buf.Bytes()))
		assert.Equal(t, http.StatusCreated, resp.StatusCode)
	})
}

func TestSetPolicies_withErroneousProvider(t *testing.T) {
	testsupport.WithSetUp(&applicationsHandlerData{}, func(data *applicationsHandlerData) {
		time.Sleep(time.Millisecond * 100)
		provider := data.providers["aName"]
		noop := provider.(*orchestrator_test.NoopProvider)
		noop.SetTestErr(errors.New("oops"))

		var buf bytes.Buffer
		policy := hexapolicy.PolicyInfo{Meta: hexapolicy.MetaInfo{Version: "v0.5"}, Actions: []hexapolicy.ActionInfo{{"anAction"}}, Subject: hexapolicy.SubjectInfo{Members: []string{"anEmail", "anotherEmail"}}, Object: hexapolicy.ObjectInfo{ResourceID: "aResourceId"}}
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
