package decisionsupport_test

import (
	"errors"
	"fmt"
	"github.com/gorilla/mux"
	"github.com/hexa-org/policy-orchestrator/pkg/decisionsupport"
	"github.com/hexa-org/policy-orchestrator/pkg/decisionsupport/providers"
	"github.com/hexa-org/policy-orchestrator/pkg/healthsupport"
	"github.com/hexa-org/policy-orchestrator/pkg/websupport"
	"github.com/stretchr/testify/assert"
	"io"
	"net"
	"net/http"
	"testing"
)

func TestMiddleware_allowed(t *testing.T) {
	provider := providers.MockDecisionProvider{Decision: true}
	support := decisionsupport.DecisionSupport{Provider: &provider, Skip: []string{"/health", "/metrics"}}

	server := startNewServer(support)
	defer websupport.Stop(server)

	// todo - fix me
	provider.On("BuildInput").Return().Times(42)
	provider.On("Allow").Return().Times(42)

	resp, _ := http.Get(fmt.Sprintf("http://%s/noop", server.Addr))
	body, _ := io.ReadAll(resp.Body)
	assert.Equal(t, http.StatusOK, resp.StatusCode)
	assert.Equal(t, "success!", string(body))
}

func TestMiddleware_notAllowed(t *testing.T) {
	provider := providers.MockDecisionProvider{Decision: false}
	support := decisionsupport.DecisionSupport{
		Provider: &provider,
		Skip:     []string{"/health", "/metrics"},
		Unauthorized: func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(401)
		},
	}

	server := startNewServer(support)
	defer websupport.Stop(server)

	provider.On("BuildInput")
	provider.On("Allow")

	resp, _ := http.Get(fmt.Sprintf("http://%s/noop", server.Addr))
	body, _ := io.ReadAll(resp.Body)
	assert.Equal(t, http.StatusUnauthorized, resp.StatusCode)
	assert.Equal(t, "", string(body))
}

func TestMiddleware_notAllowed_dueToBuildError(t *testing.T) {
	provider := providers.MockDecisionProvider{Decision: true, BuildErr: errors.New("oops")}
	support := decisionsupport.DecisionSupport{
		Provider: &provider,
		Skip:     []string{"/health", "/metrics"},
		Unauthorized: func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(401)
		},
	}

	server := startNewServer(support)
	defer websupport.Stop(server)

	provider.On("BuildInput")
	provider.On("Allow")

	resp, _ := http.Get(fmt.Sprintf("http://%s/noop", server.Addr))
	body, _ := io.ReadAll(resp.Body)
	assert.Equal(t, http.StatusUnauthorized, resp.StatusCode)
	assert.Equal(t, "", string(body))
}

func TestMiddleware_notAllowed_dueToAllowError(t *testing.T) {
	provider := providers.MockDecisionProvider{Decision: true, AllowErr: errors.New("oops")}
	support := decisionsupport.DecisionSupport{
		Provider: &provider,
		Skip:     []string{"/health", "/metrics"},
		Unauthorized: func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(401)
		},
	}

	server := startNewServer(support)
	defer websupport.Stop(server)

	provider.On("BuildInput")
	provider.On("Allow")

	resp, _ := http.Get(fmt.Sprintf("http://%s/noop", server.Addr))
	body, _ := io.ReadAll(resp.Body)
	assert.Equal(t, http.StatusUnauthorized, resp.StatusCode)
	assert.Equal(t, "", string(body))
}

func TestMiddleware_skips(t *testing.T) {
	provider := providers.MockDecisionProvider{}
	support := decisionsupport.DecisionSupport{Provider: &provider, Skip: []string{"/health", "/metrics"}}

	server := startNewServer(support)
	defer websupport.Stop(server)

	provider.On("BuildInput")
	provider.On("Allow")

	// todo - fix me
	_, _ = http.Get(fmt.Sprintf("http://%s/health", server.Addr))
	_, _ = http.Get(fmt.Sprintf("http://%s/metrics", server.Addr))
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
