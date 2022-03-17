package opasupport_test

import (
	"bytes"
	"errors"
	"fmt"
	"github.com/gorilla/mux"
	"github.com/hexa-org/policy-orchestrator/cmd/demo/opasupport"
	"github.com/hexa-org/policy-orchestrator/pkg/healthsupport"
	"github.com/hexa-org/policy-orchestrator/pkg/websupport"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"io"
	"io/ioutil"
	"log"
	"net"
	"net/http"
	"testing"
)

type MockClient struct {
	mock.Mock
	response []byte
	err      error
}

func (m *MockClient) Do(_ *http.Request) (*http.Response, error) {
	r := ioutil.NopCloser(bytes.NewReader(m.response))
	return &http.Response{StatusCode: 200, Body: r}, m.err
}

func unauthorized(w http.ResponseWriter, _ *http.Request) {
	w.WriteHeader(http.StatusUnauthorized)
}

func TestAllow(t *testing.T) {
	input := opasupport.OpaQuery{Input: map[string]interface{}{
		"method":     "GET",
		"path":       "/aUri",
		"principals": []interface{}{"allusers", "allauthenticatedusers", "sales@hexaindustries.io", "accounting@hexaindustries.io"},
	}}
	mockClient := new(MockClient)
	mockClient.response = []byte("{\"result\":true}")

	support := opasupport.NewOpaSupport(mockClient, "aUrl", unauthorized)
	allow, err := support.Allow(input)
	assert.NoError(t, err)
	assert.True(t, allow)
}

func TestAllow_bad_json(t *testing.T) {
	mockClient := &MockClient{err: errors.New("oops")}
	mockClient.response = []byte("{\"result\":true}")

	support := opasupport.NewOpaSupport(mockClient, "aUrl", unauthorized)
	allow, err := support.Allow(nil)
	assert.Error(t, err)
	assert.False(t, allow)
}

func TestNotAllow(t *testing.T) {
	input := opasupport.OpaQuery{Input: map[string]interface{}{
		"method":     "GET",
		"path":       "/aUri",
		"principals": []interface{}{"allusers"},
	}}
	mockClient := new(MockClient)
	mockClient.response = []byte("{\"result\":false}")

	support := opasupport.NewOpaSupport(mockClient, "aUrl", unauthorized)
	allow, err := support.Allow(input)
	assert.NoError(t, err)
	assert.False(t, allow)
}

func TestMiddleware(t *testing.T) {
	mockClient := new(MockClient)
	mockClient.response = []byte("{\"result\":true}")
	server := setupWithMockClient(mockClient)

	resp, err := http.Get(fmt.Sprintf("http://%s/", server.Addr))
	if err != nil {
		log.Fatalln(err)
	}
	body, _ := io.ReadAll(resp.Body)
	assert.Equal(t, "opa!", string(body))
	assert.Equal(t, http.StatusOK, resp.StatusCode)

	websupport.Stop(server)
}

func TestMiddlewareNotAllowed(t *testing.T) {
	mockClient := new(MockClient)
	mockClient.response = []byte("{\"result\":false}")
	server := setupWithMockClient(mockClient)

	resp, err := http.Get(fmt.Sprintf("http://%s/", server.Addr))
	if err != nil {
		log.Fatalln(err)
	}
	assert.Equal(t, http.StatusUnauthorized, resp.StatusCode)

	websupport.Stop(server)
}

func setupWithMockClient(mockClient *MockClient) *http.Server {
	support := opasupport.NewOpaSupport(mockClient, "aUrl", unauthorized)

	listener, _ := net.Listen("tcp", "localhost:0")
	server := websupport.Create(listener.Addr().String(), func(router *mux.Router) {
		router.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
			_, _ = w.Write([]byte("opa!"))
		})
	}, websupport.Options{})
	router := server.Handler.(*mux.Router)
	router.Use(support.Middleware)

	go websupport.Start(server, listener)

	healthsupport.WaitForHealthy(server)
	return server
}
