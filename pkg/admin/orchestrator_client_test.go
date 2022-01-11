package admin_test

import (
	"bytes"
	"encoding/base64"
	"errors"
	"fmt"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"hexa/pkg/admin"
	"io/ioutil"
	"net/http"

	"testing"
)

type MockClient struct {
	mock.Mock
	response []byte
	status   int
	err      error
}

func (m *MockClient) Do(req *http.Request) (*http.Response, error) {
	r := ioutil.NopCloser(bytes.NewReader(m.response))
	return &http.Response{StatusCode: 200, Body: r}, m.err
}

func (m *MockClient) Get(url string) (resp *http.Response, err error) {
	r := ioutil.NopCloser(bytes.NewReader(m.response))
	return &http.Response{StatusCode: 200, Body: r}, m.err
}

func TestOrchestratorClient_Health(t *testing.T) {
	mockClient := new(MockClient)
	mockClient.response = []byte("{\"status\": \"pass\"}")

	client := admin.NewOrchestratorClient(mockClient, "aKey")
	resp, _ := client.Health("localhost:8883/health")
	assert.Equal(t, "{\"status\": \"pass\"}", resp)
}

func TestOrchestratorClient_NotHealthy(t *testing.T) {
	mockClient := new(MockClient)
	mockClient.err = errors.New("anError")

	client := admin.NewOrchestratorClient(mockClient, "aKey")
	resp, _ := client.Health("localhost:8883/health")
	assert.Equal(t, "{\"status\": \"fail\"}", resp)
}

func TestOrchestratorClient_Applications(t *testing.T) {
	mockClient := new(MockClient)
	mockClient.response = []byte("{\"applications\":[{\"id\":\"anId\", \"integration_id\":\"anIntegrationId\", \"object_id\":\"anObjectId\", \"name\":\"anApp\", \"description\":\"aDescription\"}]}")
	client := admin.NewOrchestratorClient(mockClient, "aKey")

	resp, _ := client.Applications("localhost:8883/applications")
	assert.Equal(t, []admin.Application{{ID: "anId", IntegrationId: "anIntegrationId", ObjectId: "anObjectId", Name: "anApp", Description: "aDescription"}}, resp)
}

func TestOrchestratorClient_Applications_bad_get(t *testing.T) {
	mockClient := new(MockClient)
	mockClient.err = errors.New("oops")
	client := admin.NewOrchestratorClient(mockClient, "aKey")

	_, err := client.Applications("localhost:8883/applications")
	assert.Error(t, err)
}

func TestOrchestratorClient_Applications_bad_json(t *testing.T) {
	mockClient := new(MockClient)
	mockClient.response = []byte("{\"_applications\":[}")
	client := admin.NewOrchestratorClient(mockClient, "aKey")

	resp, err := client.Applications("localhost:8883/applications")
	assert.Error(t, err)
	assert.Equal(t, []admin.Application(nil), resp)
}

func TestOrchestratorClient_Integrations(t *testing.T) {
	mockClient := new(MockClient)
	key := base64.StdEncoding.EncodeToString([]byte("anotherKey"))
	mockClient.response = []byte(fmt.Sprintf("{\"integrations\":[{\"provider\":\"google\", \"key\":\"%s\"}]}", key))
	client := admin.NewOrchestratorClient(mockClient, "aKey")

	resp, _ := client.Integrations("localhost:8883/applications")
	assert.Equal(t, []admin.Integration{{Provider: "google", Key: []byte("anotherKey")}}, resp)
}

func TestOrchestratorClient_Integrations_bad_get(t *testing.T) {
	mockClient := new(MockClient)
	mockClient.err = errors.New("oops")
	client := admin.NewOrchestratorClient(mockClient, "aKey")

	_, err := client.Integrations("localhost:8883/integrations")
	assert.Error(t, err)
}

func TestOrchestratorClient_Integrations_bad_json(t *testing.T) {
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

	err := client.CreateIntegration("localhost:8883/integrations", "", []byte{})
	assert.NoError(t, err)
}

func TestOrchestratorClient_DeleteIntegrations(t *testing.T) {
	mockClient := new(MockClient)
	mockClient.status = http.StatusOK
	client := admin.NewOrchestratorClient(mockClient, "aKey")

	err := client.DeleteIntegration("localhost:8883/integrations/101")
	assert.NoError(t, err)
}
