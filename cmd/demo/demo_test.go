package main

import (
	"github.com/hexa-org/policy-orchestrator/pkg/web_support"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"io/ioutil"
	"net/http"
	"path/filepath"
	"runtime"
	"strings"
	"testing"
)

type MockClient struct {
	mock.Mock
	response string
}

func (m *MockClient) Do(req *http.Request) (*http.Response, error) {
	return &http.Response{StatusCode: 200, Body: ioutil.NopCloser(strings.NewReader(m.response))}, nil
}

func TestApp(t *testing.T) {
	client := new(MockClient)
	client.response = "{\"result\":true}"
	app := setup(client)
	assert.Equal(t, http.StatusOK, must("http://localhost:8883/health").StatusCode)
	assert.Equal(t, http.StatusOK, must("http://localhost:8883/").StatusCode)
	assert.Equal(t, http.StatusOK, must("http://localhost:8883/sales").StatusCode)
	assert.Equal(t, http.StatusOK, must("http://localhost:8883/accounting").StatusCode)
	assert.Equal(t, http.StatusOK, must("http://localhost:8883/humanresources").StatusCode)
	web_support.Stop(app)
}

func TestConfig(t *testing.T) {
	newApp()
}

func TestApp_unauthorized(t *testing.T) {
	client := new(MockClient)
	client.response = "{\"result\":false}"
	app := setup(client)
	assert.Equal(t, http.StatusUnauthorized, must("http://localhost:8883/humanresources").StatusCode)
	web_support.Stop(app)
}

func setup(client *MockClient) *http.Server {
	_, file, _, _ := runtime.Caller(0)
	resourcesDirectory := filepath.Join(file, "../../../cmd/demo/resources")
	app := App(client, "http://0.0.0.0:8887/v1/data/authz/allow", "localhost:8883", resourcesDirectory)
	go func() {
		web_support.Start(app)
	}()
	web_support.WaitForHealthy(app)
	return app
}

func must(url string) *http.Response {
	resp, _ := http.Get(url)
	return resp
}
