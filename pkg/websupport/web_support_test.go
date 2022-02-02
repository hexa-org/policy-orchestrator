package websupport_test

import (
	"fmt"
	"github.com/gorilla/mux"
	"github.com/hexa-org/policy-orchestrator/pkg/healthsupport"
	"github.com/hexa-org/policy-orchestrator/pkg/websupport"
	"github.com/stretchr/testify/assert"
	"io"
	"net"
	"net/http"
	"testing"
)

func TestServer(t *testing.T) {
	listener, _ := net.Listen("tcp", "localhost:0")
	server := websupport.Create(listener.Addr().String(), func(x *mux.Router) {}, websupport.Options{})
	go websupport.Start(server, listener)

	healthsupport.WaitForHealthy(server)

	resp, _ := http.Get(fmt.Sprintf("http://%s/health", server.Addr))
	body, _ := io.ReadAll(resp.Body)
	assert.Equal(t, "{\"status\":\"pass\"}", string(body))

	resp, _ = http.Get(fmt.Sprintf("http://%s/metrics", server.Addr))
	body, _ = io.ReadAll(resp.Body)
	assert.Contains(t, string(body), "TYPE any_request_duration_seconds histogram")

	websupport.Stop(server)
}

func TestPaths(t *testing.T) {
	listener, _ := net.Listen("tcp", "localhost:0")
	server := websupport.Create(listener.Addr().String(), func(x *mux.Router) {}, websupport.Options{})

	assert.Equal(t, 2, len(websupport.Paths(server.Handler.(*mux.Router))))
}
