package admin_test

import (
	"bytes"
	"errors"
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

func TestOrchestratorClient_ApplicationsApplications(t *testing.T) {
	mockClient := new(MockClient)
	mockClient.response = []byte("{\"applications\":[{\"name\":\"anApp\"}]}")
	client := admin.NewOrchestratorClient(mockClient, "aKey")

	resp, _ := client.Applications("localhost:8883/applications")
	assert.Equal(t, []admin.Application{{Name: "anApp"}}, resp)
}
