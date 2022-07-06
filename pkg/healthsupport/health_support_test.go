package healthsupport_test

import (
	"context"
	"fmt"
	"github.com/gorilla/mux"
	"github.com/hexa-org/policy-orchestrator/pkg/healthsupport"
	"github.com/stretchr/testify/assert"
	"io"
	"net"
	"net/http"
	"testing"
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
