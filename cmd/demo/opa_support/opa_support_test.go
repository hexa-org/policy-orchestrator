package opa_support_test

import (
	"bytes"
	"github.com/gorilla/mux"
	"github.com/hexa-org/policy-orchestrator/cmd/demo/opa_support"
	"github.com/hexa-org/policy-orchestrator/pkg/web_support"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"io"
	"io/ioutil"
	"log"
	"net/http"
	"testing"
)

type MockClient struct {
	mock.Mock
	response []byte
}

func (m *MockClient) Do(req *http.Request) (*http.Response, error) {
	r := ioutil.NopCloser(bytes.NewReader(m.response))
	return &http.Response{StatusCode: 200, Body: r}, nil
}

///

func TestAllow(t *testing.T) {
	input := opa_support.OpaQuery{Input: map[string]interface{}{
		"method":     "GET",
		"path":       "/aUri",
		"principals": []interface{}{"allusers", "allauthenticatedusers", "sales@", "accounting@"},
	}}
	mockClient := new(MockClient)
	mockClient.response = []byte("{\"result\":true}")

	support, err := opa_support.NewOpaSupport(mockClient, "aUrl")
	allow, err := support.Allow(input)
	assert.NoError(t, err)
	assert.True(t, allow)
}

func TestNotAllow(t *testing.T) {
	input := opa_support.OpaQuery{Input: map[string]interface{}{
		"method":     "GET",
		"path":       "/aUri",
		"principals": []interface{}{"allusers"},
	}}
	mockClient := new(MockClient)
	mockClient.response = []byte("{\"result\":false}")

	support, err := opa_support.NewOpaSupport(mockClient, "aUrl")
	allow, err := support.Allow(input)
	assert.NoError(t, err)
	assert.False(t, allow)
}

func TestMiddleware(t *testing.T) {
	mockClient := new(MockClient)
	mockClient.response = []byte("{\"result\":true}")
	err, server := setup(mockClient)

	resp, err := http.Get("http://localhost:8883/")
	if err != nil {
		log.Fatalln(err)
	}
	body, _ := io.ReadAll(resp.Body)
	assert.Equal(t, "opa!", string(body))
	assert.Equal(t, http.StatusOK, resp.StatusCode)

	web_support.Stop(server)
}

func TestMiddlewareNotAllowed(t *testing.T) {
	mockClient := new(MockClient)
	mockClient.response = []byte("{\"result\":false}")
	err, server := setup(mockClient)

	resp, err := http.Get("http://localhost:8883/")
	if err != nil {
		log.Fatalln(err)
	}
	assert.Equal(t, http.StatusUnauthorized, resp.StatusCode)

	web_support.Stop(server)
}

func setup(mockClient *MockClient) (error, *http.Server) {
	support, err := opa_support.NewOpaSupport(mockClient, "aUrl")

	handler := opa_support.OpaMiddleware(support, func(w http.ResponseWriter, r *http.Request) {
		_, _ = w.Write([]byte("opa!"))
	}, func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusUnauthorized)
	})

	server := web_support.Create("localhost:8883", func(router *mux.Router) {
		router.HandleFunc("/", handler).Methods("GET")
	}, web_support.Options{})

	go web_support.Start(server)

	web_support.WaitForHealthy(server)
	return err, server
}
