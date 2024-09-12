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
	"github.com/hexa-org/policy-mapper/pkg/mockOidcSupport"
	"github.com/hexa-org/policy-mapper/pkg/oauth2support"
	"github.com/hexa-org/policy-mapper/sdk"
	"github.com/hexa-org/policy-orchestrator/demo/internal/orchestrator"
	"github.com/hexa-org/policy-orchestrator/demo/internal/orchestrator/test"
	"github.com/hexa-org/policy-orchestrator/demo/pkg/dataConfigGateway"

	"github.com/hexa-org/policy-mapper/pkg/healthsupport"
	"github.com/hexa-org/policy-mapper/pkg/websupport"
	"github.com/hexa-org/policy-orchestrator/demo/pkg/testsupport"
	"github.com/stretchr/testify/assert"
)

type orchestrationHandlerData struct {
	server    *http.Server
	key       string
	providers map[string]policyprovider.Provider

	fromApp        string
	toApp          string
	toAppDifferent string

	testDir    string
	Data       *dataConfigGateway.ConfigData
	appGateway dataConfigGateway.ApplicationsDataGateway

	MockOauth       *mockOidcSupport.MockAuthServer
	oauthHttpClient *http.Client
}

func (data *orchestrationHandlerData) SetUp() {
	/*
	       data.db, _ = databasesupport.Open("postgres://orchestrator:orchestrator@localhost:5432/orchestrator_test?sslmode=disable")
	       _, _ = data.db.Exec(`
	   delete from applications;
	   delete from integrations;
	   insert into integrations (id, name, provider, key) values ('50e00619-9f15-4e85-a7e9-f26d87ea12e7', 'aName', 'noop', 'aKey');
	   insert into applications (id, integration_id, object_id, name, description) values ('6409776a-367a-483a-a194-5ccf9c4ff210', '50e00619-9f15-4e85-a7e9-f26d87ea12e7', 'anObjectId', 'aName', 'aDescription');
	   insert into applications (id, integration_id, object_id, name, description) values ('6409776a-367a-483a-a194-5ccf9c4ff211', '50e00619-9f15-4e85-a7e9-f26d87ea12e7', 'anotherObjectId', 'anotherName', 'anotherDescription');

	   insert into integrations (id, name, provider, key) values ('50e00619-9f15-4e85-a7e9-f26d87ea12e8', 'anotherName', 'azure', 'aKey');
	   insert into applications (id, integration_id, object_id, name, description) values ('6409776a-367a-483a-a194-5ccf9c4ff212', '50e00619-9f15-4e85-a7e9-f26d87ea12e8', 'andAnotherObjectId', 'andAnotherName', 'andAnotherDescription');
	   `)

	*/
	data.MockOauth = mockOidcSupport.NewMockAuthServer("clientId", "secret", map[string]interface{}{})
	mockUrlJwks, _ := url.JoinPath(data.MockOauth.Server.URL, "/jwks")
	// Set Env for Jwt Token Validation by Orchestrator handlers
	_ = os.Setenv(oauth2support.EnvOAuthJwksUrl, mockUrlJwks)
	_ = os.Setenv(oauth2support.EnvJwtRealm, "TEST_REALM")
	_ = os.Setenv(oauth2support.EnvJwtAuth, "true")

	// Set Env for Orchestrator Client
	_ = os.Setenv(oauth2support.EnvOAuthClientId, "clientId")
	_ = os.Setenv(oauth2support.EnvOAuthClientSecret, "secret")

	_ = os.Setenv(oauth2support.EnvOAuthTokenEndpoint, fmt.Sprintf("%s/token", data.MockOauth.Server.URL))

	data.oauthHttpClient = oauth2support.NewJwtClientHandler().GetHttpClient()

	tempDir, _ := os.MkdirTemp("", "hexa-orchestrator-*")

	data.testDir = tempDir

	testConfigPath := filepath.Join(data.testDir, ".hexa", "config.json")
	_ = os.Setenv(sdk.EnvTestProvider, "noop")
	// _ = os.Unsetenv(sdk.EnvTestProvider)

	_ = os.Setenv(sdk.EnvTestProvider, sdk.ProviderTypeMock)
	_ = os.Setenv(dataConfigGateway.EnvIntegrationConfigFile, testConfigPath)

	data.Data, _ = dataConfigGateway.NewIntegrationConfigData()
	data.appGateway = data.Data.GetApplicationDataGateway()

	_, _ = data.Data.Create("50e00619-9f15-4e85-a7e9-f26d87ea12e7", "noop", []byte("aKey"))
	integration := data.Data.Integrations["50e00619-9f15-4e85-a7e9-f26d87ea12e7"]
	apps := []policyprovider.ApplicationInfo{
		{"anObjectId", "aName", "aDescription", "aService"},
		{"anotherObjectId", "anotherName", "anotherDescription", "anotherService"},
	}
	integration.Apps["6409776a-367a-483a-a194-5ccf9c4ff210"] = apps[0]
	integration.Apps["6409776a-367a-483a-a194-5ccf9c4ff211"] = apps[0]

	_, _ = data.Data.Create("50e00619-9f15-4e85-a7e9-f26d87ea12e8", "noop", []byte("aKey"))
	integration = data.Data.Integrations["50e00619-9f15-4e85-a7e9-f26d87ea12e8"]
	integration.Apps["6409776a-367a-483a-a194-5ccf9c4ff212"] = policyprovider.ApplicationInfo{"andAnotherObjectId", "andAnotherName", "andAnotherDescription", "andAnotherService"}
	_, _ = data.Data.Create("50e00619-9f15-4e85-a7e9-f26d87ea12e9", "noop", []byte("aKey"))
	integration = data.Data.Integrations["50e00619-9f15-4e85-a7e9-f26d87ea12e8"]
	integration.Apps["6409776a-367a-483a-a194-5ccf9c4ff213"] = policyprovider.ApplicationInfo{"yetAnotherObjectId", "yetAnotherName", "yetAnotherDescription", "yetAnotherService"}

	data.fromApp = "6409776a-367a-483a-a194-5ccf9c4ff210"
	data.toApp = "6409776a-367a-483a-a194-5ccf9c4ff211"
	data.toAppDifferent = "6409776a-367a-483a-a194-5ccf9c4ff212"

	listener, _ := net.Listen("tcp", "localhost:0")
	addr := listener.Addr().String()

	hash := sha256.Sum256([]byte("aKey"))
	data.key = hex.EncodeToString(hash[:])

	data.providers = make(map[string]policyprovider.Provider)
	data.providers["50e00619-9f15-4e85-a7e9-f26d87ea12e7"] = &orchestratorNoopProvider.NoopProvider{}
	// data.providers["azure"] = microsoftazure.NewAzureProvider() This will auto load
	handlers := orchestrator.LoadHandlers(data.Data, data.providers)
	data.server = websupport.Create(addr, handlers, websupport.Options{})
	go websupport.Start(data.server, listener)
	healthsupport.WaitForHealthy(data.server)
}

func (data *orchestrationHandlerData) TearDown() {
	// _ = data.db.Close()
	websupport.Stop(data.server)
	data.oauthHttpClient.CloseIdleConnections()
	data.MockOauth.Shutdown()
	data.MockOauth = nil
	data.server = nil
	data.Data = nil
	data.appGateway = nil
	data.oauthHttpClient = nil
	_ = os.RemoveAll(data.testDir)
	data.testDir = ""
}

func TestOrchestration(t *testing.T) {
	testsupport.WithSetUp(&orchestrationHandlerData{}, func(data *orchestrationHandlerData) {
		url := fmt.Sprintf("http://%s/orchestration", data.server.Addr)
		marshal, _ := json.Marshal(orchestrator.Orchestration{From: data.fromApp, To: data.toApp})

		resp, err := data.oauthHttpClient.Post(url, "application/json", bytes.NewReader(marshal))
		if err != nil {
			fmt.Println(err.Error())
		}
		assert.Equal(t, http.StatusCreated, resp.StatusCode)
	})
}

func TestOrchestration_failsAcrossProviders(t *testing.T) {
	testsupport.WithSetUp(&orchestrationHandlerData{}, func(data *orchestrationHandlerData) {
		url := fmt.Sprintf("http://%s/orchestration", data.server.Addr)
		marshal, _ := json.Marshal(orchestrator.Orchestration{From: data.fromApp, To: data.toAppDifferent})

		resp, err := data.oauthHttpClient.Post(url, "application/json", bytes.NewReader(marshal))
		assert.Nil(t, err)
		assert.Equal(t, http.StatusInternalServerError, resp.StatusCode)
	})
}
