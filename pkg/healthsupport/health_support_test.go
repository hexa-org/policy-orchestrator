package healthsupport_test

import (
	"context"
	"fmt"
	"io"
	"net"
	"net/http"
	"testing"

	"github.com/gorilla/mux"
	"github.com/hexa-org/policy-orchestrator/pkg/healthsupport"
	"github.com/stretchr/testify/assert"
)

func TestHealth(t *testing.T) {
	listener, _ := net.Listen("tcp", "localhost:0")
	router := mux.NewRouter()
	router.HandleFunc("/health", healthsupport.HealthHandlerFunction).Methods("GET")
	server := &http.Server{
		Addr:    listener.Addr().String(),
		Handler: router,
	}
	go func() {
		_ = server.Serve(listener)
	}()
	healthsupport.WaitForHealthy(server)

	resp, _ := http.Get(fmt.Sprintf("http://%s/health", server.Addr))
	body, _ := io.ReadAll(resp.Body)
	assert.Equal(t, "{\"status\":\"pass\"}", string(body))
	_ = server.Shutdown(context.Background())
}

func TestWaitForHealthyWithClient(t *testing.T) {
	listener, _ := net.Listen("tcp", "localhost:0")
	router := mux.NewRouter()
	router.HandleFunc("/health", healthsupport.HealthHandlerFunction).Methods("GET")
	server := &http.Server{
		Addr:    listener.Addr().String(),
		Handler: router,
	}
	go func() {
		_ = server.Serve(listener)
	}()
	defer server.Shutdown(context.Background())

	client := &mockClient{}
	healthsupport.WaitForHealthyWithClient(server, client, fmt.Sprintf("http://%s/health", server.Addr))

	resp, _ := http.Get(fmt.Sprintf("http://%s/health", server.Addr))
	body, _ := io.ReadAll(resp.Body)
	assert.Equal(t, "{\"status\":\"pass\"}", string(body))
	assert.True(t, client.getCalled)
}

type mockClient struct {
	getCalled bool
}

func (m *mockClient) Get(_ string) (*http.Response, error) {
	m.getCalled = true
	return &http.Response{StatusCode: http.StatusOK}, nil
}
