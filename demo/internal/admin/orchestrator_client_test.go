package admin_test

import (
	"bytes"
	"encoding/base64"
	"errors"
	"fmt"
	"io"
	"log"
	"net/http"
	"net/url"
	"os"
	"testing"

	"github.com/hexa-org/policy-mapper/pkg/hexapolicy"
	"github.com/hexa-org/policy-mapper/pkg/mockOidcSupport"
	"github.com/hexa-org/policy-mapper/pkg/oauth2support"
	"github.com/hexa-org/policy-orchestrator/demo/internal/admin"
	"golang.org/x/oauth2"

	"github.com/go-playground/validator/v10"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

type MockClient struct {
	mock.Mock
	response []byte
	status   int
	err      error
}

func (m *MockClient) Do(_ *http.Request) (*http.Response, error) {
	r := io.NopCloser(bytes.NewReader(m.response))
	return &http.Response{StatusCode: m.status, Body: r}, m.err
}

func (m *MockClient) Get(_ string) (resp *http.Response, err error) {
	r := io.NopCloser(bytes.NewReader(m.response))
	return &http.Response{StatusCode: m.status, Body: r}, m.err
}

func TestOrchestratorJwtClient(t *testing.T) {
	log.Println("Starting Mock OAuth Server")
	cid := "testClientId"
	mockAuth := mockOidcSupport.NewMockAuthServer(cid, "secret", map[string]interface{}{})
	defer mockAuth.Server.Close()
	mockerAddr := mockAuth.Server.URL
	mockUrlJwks, _ := url.JoinPath(mockerAddr, "/jwks")
	assert.NotEmpty(t, mockUrlJwks)
	_ = os.Setenv(oauth2support.EnvOAuthJwksUrl, mockUrlJwks)
	_ = os.Setenv(oauth2support.EnvJwtRealm, "TEST_REALM")
	_ = os.Setenv(oauth2support.EnvJwtAuth, "true")

	client := admin.NewOrchestratorClient(nil, "localhost:8883")
	httpClient := client.GetHttpClient()
	switch h := httpClient.(type) {
	case *http.Client:
		switch h.Transport.(type) {
		case *oauth2.Transport:
			fmt.Println("Correct Jwt configured")
		default:
			assert.Fail(t, "Wrong transport type")
		}
	default:
		assert.Fail(t, "Unexpected HTTP Client")
	}
}

func TestOrchestratorClient_Health(t *testing.T) {
	mockClient := new(MockClient)
	mockClient.response = []byte("[{\"name\":\"noop\",\"pass\":\"true\"}]")
	client := admin.NewOrchestratorClient(mockClient, "localhost:8883")

	resp, _ := client.Health()
	assert.Equal(t, "[{\"name\":\"noop\",\"pass\":\"true\"}]", resp)
}

func TestOrchestratorClient_NotHealthy(t *testing.T) {
	mockClient := new(MockClient)
	mockClient.err = errors.New("anError")
	client := admin.NewOrchestratorClient(mockClient, "localhost:8883")

	resp, _ := client.Health()
	assert.Equal(t, "[{\"name\":\"Unreachable\",\"pass\":\"fail\"}]", resp)
}

func TestOrchestratorClient_Applications(t *testing.T) {
	mockClient := new(MockClient)
	mockClient.response = []byte("{\"applications\":[{\"id\":\"anId\", \"integration_id\":\"anIntegrationId\", \"object_id\":\"anObjectId\", \"name\":\"anApp\", \"description\":\"aDescription\", \"provider_name\":\"aProviderName\", \"service\":\"aService\"}]}")
	mockClient.status = http.StatusOK
	client := admin.NewOrchestratorClient(mockClient, "localhost:8883")

	resp, _ := client.Applications(false)
	assert.Equal(t, []admin.Application{{ID: "anId", IntegrationId: "anIntegrationId", ObjectId: "anObjectId", Name: "anApp", Description: "aDescription", ProviderName: "aProviderName", Service: "aService"}}, resp)
}

func TestOrchestratorClient_Applications_withErroneousGet(t *testing.T) {
	mockClient := new(MockClient)
	mockClient.err = errors.New("oops")
	client := admin.NewOrchestratorClient(mockClient, "localhost:8883")

	_, err := client.Applications(false)
	assert.Error(t, err)
}

func TestOrchestratorClient_Applications_withBadJson(t *testing.T) {
	mockClient := new(MockClient)
	mockClient.response = []byte("{\"_applications\":[}")
	client := admin.NewOrchestratorClient(mockClient, "localhost:8883")

	resp, err := client.Applications(false)
	assert.Error(t, err)
	assert.Equal(t, []admin.Application(nil), resp)
}

func TestOrchestratorClient_Application(t *testing.T) {
	mockClient := new(MockClient)
	mockClient.response = []byte("{\"id\":\"anId\", \"integration_id\":\"anIntegrationId\", \"object_id\":\"anObjectId\", \"name\":\"anApp\", \"description\":\"aDescription\", \"service\":\"aService\"}}")
	mockClient.status = http.StatusOK
	client := admin.NewOrchestratorClient(mockClient, "localhost:8883")

	resp, _ := client.Application("anId")
	assert.Equal(t, admin.Application{ID: "anId", IntegrationId: "anIntegrationId", ObjectId: "anObjectId", Name: "anApp", Description: "aDescription", Service: "aService"}, resp)
}

func TestOrchestratorClient_Application_withBadJson(t *testing.T) {
	mockClient := new(MockClient)
	mockClient.response = []byte("_{\"_id\":\"anId\"}}")
	client := admin.NewOrchestratorClient(mockClient, "localhost:8883")

	_, err := client.Application("anId")
	assert.Error(t, err)
}

func TestOrchestratorClient_Integrations(t *testing.T) {
	key := base64.StdEncoding.EncodeToString([]byte("anotherKey"))

	mockClient := new(MockClient)
	mockClient.response = []byte(fmt.Sprintf("{\"integrations\":[{\"provider\":\"google_cloud\", \"key\":\"%s\"}]}", key))
	mockClient.status = http.StatusOK
	client := admin.NewOrchestratorClient(mockClient, "localhost:8883")

	resp, _ := client.Integrations()
	assert.Equal(t, []admin.Integration{{Provider: "google_cloud", Key: []byte("anotherKey")}}, resp)
}

func TestOrchestratorClient_Integrations_withErroneousGet(t *testing.T) {
	mockClient := new(MockClient)
	mockClient.err = errors.New("oops")
	client := admin.NewOrchestratorClient(mockClient, "localhost:8883")

	_, err := client.Integrations()
	assert.Error(t, err)
}

func TestOrchestratorClient_Integrations_withBadJson(t *testing.T) {
	mockClient := new(MockClient)
	mockClient.response = []byte(fmt.Sprintf("{\"_integrations\":[}"))
	client := admin.NewOrchestratorClient(mockClient, "localhost:8883")

	resp, err := client.Integrations()
	assert.Error(t, err)
	assert.Equal(t, []admin.Integration(nil), resp)
}

func TestOrchestratorClient_CreateIntegrations(t *testing.T) {
	mockClient := new(MockClient)
	mockClient.status = http.StatusCreated
	client := admin.NewOrchestratorClient(mockClient, "localhost:8883")

	err := client.CreateIntegration("aName", "aProvider", []byte{})
	assert.NoError(t, err)
}

func TestOrchestratorClient_DeleteIntegrations(t *testing.T) {
	mockClient := new(MockClient)
	mockClient.status = http.StatusOK
	client := admin.NewOrchestratorClient(mockClient, "localhost:8883")

	err := client.DeleteIntegration("101")
	assert.NoError(t, err)
}

func TestOrchestratorClient_GetPolicy(t *testing.T) {
	mockClient := new(MockClient)
	mockClient.status = http.StatusOK
	rawJson := `{
  "policies": [
    {
      "meta": {
        "version": "aVersion"
      },
      "actions": [
        "anAction"
      ],
      "subjects": [
        "aUser"
      ],
      "object": "aResourceId"
    },
    {
      "meta": {
        "version": "anotherVersion"
      },
      "actions": [
        "anotherAction"
      ],
      "subjects": [
        "anotherUser"
      ],
      "object": "anotherResourceId"
    }
  ]
}`
	mockClient.response = []byte(rawJson)
	client := admin.NewOrchestratorClient(mockClient, "localhost:8883")

	resp, raw, _ := client.GetPolicies("anId")
	assert.Equal(t, rawJson, raw)
	assert.Equal(t, "aVersion", resp[0].Meta.Version)
	assert.Equal(t, "anAction", resp[0].Actions[0].String())
	assert.Equal(t, []string{"aUser"}, resp[0].Subjects.String())
	assert.Equal(t, "aResourceId", resp[0].Object.String())

	validate := validator.New()
	errPolicies := validate.Var(resp, "omitempty,dive")
	if errPolicies != nil {
		fmt.Println(errPolicies)
		t.Fail()
	}
}

func TestOrchestratorClient_GetPolicy_withErroneousGet(t *testing.T) {
	mockClient := new(MockClient)
	mockClient.err = errors.New("oops")
	client := admin.NewOrchestratorClient(mockClient, "localhost:8883")

	resp, _, err := client.GetPolicies("anId")
	assert.Error(t, err)
	assert.Equal(t, []hexapolicy.PolicyInfo{}, resp)
}

func TestOrchestratorClient_GetPolicy_withBadJson(t *testing.T) {
	mockClient := new(MockClient)
	mockClient.status = http.StatusOK
	client := admin.NewOrchestratorClient(mockClient, "localhost:8883")

	resp, _, err := client.GetPolicies("anId")
	assert.Error(t, err)
	assert.Equal(t, []hexapolicy.PolicyInfo{}, resp)
}

func TestOrchestratorClient_SetPolicy(t *testing.T) {
	mockClient := new(MockClient)
	mockClient.status = http.StatusCreated

	// policies := "{\"policies\":[[{\"version\":\"aVersion\",\"action_uri\":\"anAction\",\"subject\":{\"members\":[\"aUser\"]},\"object\":{\"resource_id\":\"aResourceId\"}},{\"version\":\"aVersion\",\"action\":\"anotherAction\",\"subject\":{\"members\":[\"anotherUser\"]},\"object\":{\"resource_id\":\"anotherResourceId\"}}]}"
	policies := `{
  "policies": [
    {
      "meta": {
        "version": "aVersion"
      },
      "actions": [
        "anAction"
      ],
      "subjects": [
        "aUser"
      ],
      "object": "aResourceId"
    },
    {
      "meta": {
        "version": "anotherVersion"
      },
      "actions": [
        "anotherAction"
      ],
      "subjects": [
        "anotherUser"
      ],
      "object": "anotherResourceId"
    }
  ]
}`
	client := admin.NewOrchestratorClient(mockClient, "localhost:8883")
	err := client.SetPolicies("anId", policies)
	assert.NoError(t, err)
}

func TestOrchestratorClient_SetPolicy_withErroneousSet(t *testing.T) {
	mockClient := new(MockClient)
	mockClient.err = errors.New("oops")
	client := admin.NewOrchestratorClient(mockClient, "localhost:8883")
	err := client.SetPolicies("anId", "")
	assert.Error(t, err)
}

func TestOrchestratorClient_SetPolicy_withBadStatus(t *testing.T) {
	mockClient := new(MockClient)
	mockClient.status = 500
	mockClient.response = []byte("shoot")
	client := admin.NewOrchestratorClient(mockClient, "localhost:8883")
	err := client.SetPolicies("anId", "")
	assert.Error(t, err)
	assert.Equal(t, "shoot", err.Error())
}

func TestOrchestrationClient_Orchestration(t *testing.T) {
	mockClient := new(MockClient)
	mockClient.status = http.StatusCreated
	client := admin.NewOrchestratorClient(mockClient, "localhost:8883")
	err := client.Orchestration("fromId", "toId")
	assert.NoError(t, err)
}

func TestOrchestrationClient_Orchestration_withError(t *testing.T) {
	mockClient := new(MockClient)
	mockClient.err = errors.New("oops")
	client := admin.NewOrchestratorClient(mockClient, "localhost:8883")
	err := client.Orchestration("fromId", "toId")
	assert.Error(t, err)
}
