package main

import (
	"fmt"
	"github.com/hexa-org/policy-orchestrator/pkg/healthsupport"
	"github.com/hexa-org/policy-orchestrator/pkg/websupport"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"io/ioutil"
	"net"
	"net/http"
	"os"
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
	assert.Equal(t, http.StatusOK, must(fmt.Sprintf("http://%s/health", app.Addr)).StatusCode)
	assert.Equal(t, http.StatusOK, must(fmt.Sprintf("http://%s/", app.Addr)).StatusCode)
	assert.Equal(t, http.StatusOK, must(fmt.Sprintf("http://%s/sales", app.Addr)).StatusCode)
	assert.Equal(t, http.StatusOK, must(fmt.Sprintf("http://%s/accounting", app.Addr)).StatusCode)
	assert.Equal(t, http.StatusOK, must(fmt.Sprintf("http://%s/humanresources", app.Addr)).StatusCode)
	websupport.Stop(app)
}

func TestConfigWithPort(t *testing.T) {
	_ = os.Setenv("PORT", "0")
	_ = os.Setenv("OPA_SERVER_URL", "http://localhost:8887/v1/data/authz/allow")
	newApp("localhost:0")
}

func TestApp_unauthorized(t *testing.T) {
	client := new(MockClient)
	client.response = "{\"result\":false}"
	app := setup(client)
	assert.Equal(t, http.StatusUnauthorized, must(fmt.Sprintf("http://%s/humanresources", app.Addr)).StatusCode)
	websupport.Stop(app)
}

func setup(client *MockClient) *http.Server {
	_, file, _, _ := runtime.Caller(0)
	resourcesDirectory := filepath.Join(file, "../../../cmd/demo/resources")
	listener, _ := net.Listen("tcp", "localhost:0")
	app := App(client, "http://localhost:8887/v1/data/authz/allow", listener.Addr().String(), resourcesDirectory)
	go func() {
		websupport.Start(app, listener)
	}()
	healthsupport.WaitForHealthy(app)
	return app
}

func must(url string) *http.Response {
	resp, _ := http.Get(url)
	return resp
}
