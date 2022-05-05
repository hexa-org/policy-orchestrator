package decisionsupport_test

import (
	"errors"
	"fmt"
	"github.com/gorilla/mux"
	"github.com/hexa-org/policy-orchestrator/pkg/decisionsupport"
	"github.com/hexa-org/policy-orchestrator/pkg/decisionsupportproviders"
	"github.com/hexa-org/policy-orchestrator/pkg/healthsupport"
	"github.com/hexa-org/policy-orchestrator/pkg/websupport"
	"github.com/stretchr/testify/assert"
	"io"
	"net"
	"net/http"
	"testing"
)

func TestMiddleware_allowed(t *testing.T) {
	provider := decisionsupportproviders.MockDecisionProvider{Decision: true}
	support := decisionsupport.DecisionSupport{Provider: &provider, Skip: []string{"/health", "/metrics"}}
	provider.On("BuildInput").Once()
	provider.On("Allow").Once()

	server := startNewServer(support)
	defer websupport.Stop(server)

	resp, _ := http.Get(fmt.Sprintf("http://%s/noop", server.Addr))
	body, _ := io.ReadAll(resp.Body)
	assert.Equal(t, http.StatusOK, resp.StatusCode)
	assert.Equal(t, "success!", string(body))
	provider.AssertExpectations(t)
}

func TestMiddleware_notAllowed(t *testing.T) {
	provider := decisionsupportproviders.MockDecisionProvider{Decision: false}
	provider.On("BuildInput").Once()
	provider.On("Allow").Once()

	support := decisionsupport.DecisionSupport{
		Provider: &provider,
		Skip:     []string{"/health", "/metrics"},
		Unauthorized: func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(401)
		},
	}

	server := startNewServer(support)
	defer websupport.Stop(server)

	resp, _ := http.Get(fmt.Sprintf("http://%s/noop", server.Addr))
	body, _ := io.ReadAll(resp.Body)
	assert.Equal(t, http.StatusUnauthorized, resp.StatusCode)
	assert.Equal(t, "", string(body))
	provider.AssertExpectations(t)
}

func TestMiddleware_notAllowed_dueToBuildError(t *testing.T) {
	provider := decisionsupportproviders.MockDecisionProvider{Decision: true, BuildErr: errors.New("oops")}
	provider.On("BuildInput").Once()

	support := decisionsupport.DecisionSupport{
		Provider: &provider,
		Skip:     []string{"/health", "/metrics"},
		Unauthorized: func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(401)
		},
	}

	server := startNewServer(support)
	defer websupport.Stop(server)

	resp, _ := http.Get(fmt.Sprintf("http://%s/noop", server.Addr))
	body, _ := io.ReadAll(resp.Body)
	assert.Equal(t, http.StatusUnauthorized, resp.StatusCode)
	assert.Equal(t, "", string(body))
	provider.AssertExpectations(t)
	provider.AssertNotCalled(t, "Allow")
}

func TestMiddleware_notAllowed_dueToAllowError(t *testing.T) {
	provider := decisionsupportproviders.MockDecisionProvider{Decision: true, AllowErr: errors.New("oops")}
	provider.On("BuildInput").Once()
	provider.On("Allow").Once()

	support := decisionsupport.DecisionSupport{
		Provider: &provider,
		Skip:     []string{"/health", "/metrics"},
		Unauthorized: func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(401)
		},
	}

	server := startNewServer(support)
	defer websupport.Stop(server)

	resp, _ := http.Get(fmt.Sprintf("http://%s/noop", server.Addr))
	body, _ := io.ReadAll(resp.Body)
	assert.Equal(t, http.StatusUnauthorized, resp.StatusCode)
	assert.Equal(t, "", string(body))
	provider.AssertExpectations(t)
}

func TestMiddleware_skips(t *testing.T) {
	provider := decisionsupportproviders.MockDecisionProvider{}

	support := decisionsupport.DecisionSupport{Provider: &provider, Skip: []string{"/health", "/metrics"}}

	server := startNewServer(support)
	defer websupport.Stop(server)

	resp, _ := http.Get(fmt.Sprintf("http://%s/health", server.Addr))
	assert.Equal(t, http.StatusOK, resp.StatusCode)

	resp, _ = http.Get(fmt.Sprintf("http://%s/metrics", server.Addr))
	assert.Equal(t, http.StatusOK, resp.StatusCode)

	provider.AssertNotCalled(t, "BuildInput")
	provider.AssertNotCalled(t, "Allow")
}

///

func startNewServer(support decisionsupport.DecisionSupport) *http.Server {
	listener, _ := net.Listen("tcp", "localhost:0")
	server := websupport.Create(listener.Addr().String(), func(router *mux.Router) {
		router.HandleFunc("/noop", func(w http.ResponseWriter, r *http.Request) {
			_, _ = w.Write([]byte("success!"))
		})
	}, websupport.Options{})

	router := server.Handler.(*mux.Router)
	router.Use(support.Middleware)
	go websupport.Start(server, listener)
	healthsupport.WaitForHealthy(server)
	return server
}
