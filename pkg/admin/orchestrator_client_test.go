package admin_test

import (
	"bytes"
	"encoding/base64"
	"errors"
	"fmt"
	"io/ioutil"
	"net/http"
	"testing"

	"github.com/go-playground/validator/v10"
	"github.com/hexa-org/policy-orchestrator/pkg/admin"
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
	r := ioutil.NopCloser(bytes.NewReader(m.response))
	return &http.Response{StatusCode: m.status, Body: r}, m.err
}

func (m *MockClient) Get(_ string) (resp *http.Response, err error) {
	r := ioutil.NopCloser(bytes.NewReader(m.response))
	return &http.Response{StatusCode: m.status, Body: r}, m.err
}

func TestOrchestratorClient_Health(t *testing.T) {
	mockClient := new(MockClient)
	mockClient.response = []byte("[{\"name\":\"noop\",\"pass\":\"true\"}]")
	client := admin.NewOrchestratorClient(mockClient, "aKey")

	resp, _ := client.Health("localhost:8883/health")
	assert.Equal(t, "[{\"name\":\"noop\",\"pass\":\"true\"}]", resp)
}

func TestOrchestratorClient_NotHealthy(t *testing.T) {
	mockClient := new(MockClient)
	mockClient.err = errors.New("anError")
	client := admin.NewOrchestratorClient(mockClient, "aKey")

	resp, _ := client.Health("localhost:8883/health")
	assert.Equal(t, "[{\"name\":\"Unreachable\",\"pass\":\"fail\"}]", resp)
}

func TestOrchestratorClient_Applications(t *testing.T) {
	mockClient := new(MockClient)
	mockClient.response = []byte("{\"applications\":[{\"id\":\"anId\", \"integration_id\":\"anIntegrationId\", \"object_id\":\"anObjectId\", \"name\":\"anApp\", \"description\":\"aDescription\", \"provider_name\":\"aProviderName\"}]}")
	mockClient.status = http.StatusOK
	client := admin.NewOrchestratorClient(mockClient, "aKey")

	resp, _ := client.Applications("localhost:8883/applications")
	assert.Equal(t, []admin.Application{{ID: "anId", IntegrationId: "anIntegrationId", ObjectId: "anObjectId", Name: "anApp", Description: "aDescription", ProviderName: "aProviderName"}}, resp)
}

func TestOrchestratorClient_Applications_withErroneousGet(t *testing.T) {
	mockClient := new(MockClient)
	mockClient.err = errors.New("oops")
	client := admin.NewOrchestratorClient(mockClient, "aKey")

	_, err := client.Applications("localhost:8883/applications")
	assert.Error(t, err)
}

func TestOrchestratorClient_Applications_withBadJson(t *testing.T) {
	mockClient := new(MockClient)
	mockClient.response = []byte("{\"_applications\":[}")
	client := admin.NewOrchestratorClient(mockClient, "aKey")

	resp, err := client.Applications("localhost:8883/applications")
	assert.Error(t, err)
	assert.Equal(t, []admin.Application(nil), resp)
}

func TestOrchestratorClient_Application(t *testing.T) {
	mockClient := new(MockClient)
	mockClient.response = []byte("{\"id\":\"anId\", \"integration_id\":\"anIntegrationId\", \"object_id\":\"anObjectId\", \"name\":\"anApp\", \"description\":\"aDescription\"}}")
	mockClient.status = http.StatusOK
	client := admin.NewOrchestratorClient(mockClient, "aKey")

	resp, _ := client.Application("localhost:8883/applications/anId")
	assert.Equal(t, admin.Application{ID: "anId", IntegrationId: "anIntegrationId", ObjectId: "anObjectId", Name: "anApp", Description: "aDescription"}, resp)
}

func TestOrchestratorClient_Application_withBadJson(t *testing.T) {
	mockClient := new(MockClient)
	mockClient.response = []byte("_{\"_id\":\"anId\"}}")
	client := admin.NewOrchestratorClient(mockClient, "aKey")

	_, err := client.Application("localhost:8883/applications/anId")
	assert.Error(t, err)
}

func TestOrchestratorClient_Integrations(t *testing.T) {
	key := base64.StdEncoding.EncodeToString([]byte("anotherKey"))

	mockClient := new(MockClient)
	mockClient.response = []byte(fmt.Sprintf("{\"integrations\":[{\"provider\":\"google_cloud\", \"key\":\"%s\"}]}", key))
	mockClient.status = http.StatusOK
	client := admin.NewOrchestratorClient(mockClient, "aKey")

	resp, _ := client.Integrations("localhost:8883/applications")
	assert.Equal(t, []admin.Integration{{Provider: "google_cloud", Key: []byte("anotherKey")}}, resp)
}

func TestOrchestratorClient_Integrations_withErroneousGet(t *testing.T) {
	mockClient := new(MockClient)
	mockClient.err = errors.New("oops")
	client := admin.NewOrchestratorClient(mockClient, "aKey")

	_, err := client.Integrations("localhost:8883/integrations")
	assert.Error(t, err)
}

func TestOrchestratorClient_Integrations_withBadJson(t *testing.T) {
	mockClient := new(MockClient)
	mockClient.response = []byte(fmt.Sprintf("{\"_integrations\":[}"))
	client := admin.NewOrchestratorClient(mockClient, "aKey")

	resp, err := client.Integrations("localhost:8883/integrations")
	assert.Error(t, err)
	assert.Equal(t, []admin.Integration(nil), resp)
}

func TestOrchestratorClient_CreateIntegrations(t *testing.T) {
	mockClient := new(MockClient)
	mockClient.status = http.StatusCreated
	client := admin.NewOrchestratorClient(mockClient, "aKey")

	err := client.CreateIntegration("localhost:8883/integrations", "aName", "aProvider", []byte{})
	assert.NoError(t, err)
}

func TestOrchestratorClient_DeleteIntegrations(t *testing.T) {
	mockClient := new(MockClient)
	mockClient.status = http.StatusOK
	client := admin.NewOrchestratorClient(mockClient, "aKey")

	err := client.DeleteIntegration("localhost:8883/integrations/101")
	assert.NoError(t, err)
}

func TestOrchestratorClient_GetPolicy(t *testing.T) {
	mockClient := new(MockClient)
	mockClient.status = http.StatusOK
	rawJson := "{\"policies\":[" +
		"{\"meta\":{\"version\":\"aVersion\"},\"actions\":[{\"action_uri\": \"anAction\"}],\"subject\":{\"members\":[\"aUser\"]},\"object\":{\"resource_id\":\"aResourceId\"}}," +
		"{\"meta\":{\"version\":\"anotherVersion\"},\"actions\":[{\"action\": \"anotherAction\"}],\"subject\":{\"members\":[\"anotherUser\"]},\"object\":{\"resource_id\":\"anotherResourceId\"}}]}"
	mockClient.response = []byte(rawJson)
	client := admin.NewOrchestratorClient(mockClient, "aKey")

	resp, raw, _ := client.GetPolicies("localhost:8883/applications/anId/policies")
	assert.Equal(t, rawJson, raw)
	assert.Equal(t, "aVersion", resp[0].Meta.Version)
	assert.Equal(t, "anAction", resp[0].Actions[0].ActionUri)
	assert.Equal(t, []string{"aUser"}, resp[0].Subject.Members)
	assert.Equal(t, "aResourceId", resp[0].Object.ResourceID)

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
	client := admin.NewOrchestratorClient(mockClient, "aKey")

	resp, _, err := client.GetPolicies("localhost:8883/applications/anId/policies")
	assert.Error(t, err)
	assert.Equal(t, []admin.Policy{}, resp)
}

func TestOrchestratorClient_GetPolicy_withBadJson(t *testing.T) {
	mockClient := new(MockClient)
	mockClient.status = http.StatusOK
	client := admin.NewOrchestratorClient(mockClient, "aKey")

	resp, _, err := client.GetPolicies("localhost:8883/applications/anId/policies")
	assert.Error(t, err)
	assert.Equal(t, []admin.Policy{}, resp)
}

func TestOrchestratorClient_SetPolicy(t *testing.T) {
	mockClient := new(MockClient)
	mockClient.status = http.StatusCreated
	policies := "{\"policies\":[[{\"version\":\"aVersion\",\"action_uri\":\"anAction\",\"subject\":{\"members\":[\"aUser\"]},\"object\":{\"resource_id\":\"aResourceId\"}},{\"version\":\"aVersion\",\"action\":\"anotherAction\",\"subject\":{\"members\":[\"anotherUser\"]},\"object\":{\"resource_id\":\"anotherResourceId\"}}]}"
	client := admin.NewOrchestratorClient(mockClient, "aKey")
	err := client.SetPolicies("localhost:8883/applications/anId/policies", policies)
	assert.NoError(t, err)
}

func TestOrchestratorClient_SetPolicy_withErroneousSet(t *testing.T) {
	mockClient := new(MockClient)
	mockClient.err = errors.New("oops")
	client := admin.NewOrchestratorClient(mockClient, "aKey")
	err := client.SetPolicies("localhost:8883/applications/anId/policies", "")
	assert.Error(t, err)
}

func TestOrchestratorClient_SetPolicy_withBadStatus(t *testing.T) {
	mockClient := new(MockClient)
	mockClient.status = 500
	mockClient.response = []byte("shoot")
	client := admin.NewOrchestratorClient(mockClient, "aKey")
	err := client.SetPolicies("localhost:8883/applications/anId/policies", "")
	assert.Error(t, err)
	assert.Equal(t, "shoot", err.Error())
}
