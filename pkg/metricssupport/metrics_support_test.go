package metricssupport_test

import (
	"context"
	"fmt"
	"io"
	"net"
	"net/http"
	"testing"

	"github.com/gorilla/mux"
	"github.com/hexa-org/policy-orchestrator/pkg/healthsupport"
	"github.com/hexa-org/policy-orchestrator/pkg/metricssupport"
	"github.com/stretchr/testify/assert"
)

func TestMetrics(t *testing.T) {
	listener, _ := net.Listen("tcp", "localhost:0")
	router := mux.NewRouter()
	router.Use(metricssupport.MetricsMiddleware)
	router.Path("/metrics").Handler(metricssupport.MetricsHandler())

	// note - needed for health check below
	router.HandleFunc("/health", healthsupport.HealthHandlerFunction).Methods("GET")

	server := &http.Server{
		Addr:    listener.Addr().String(),
		Handler: router,
	}
	go func() {
		_ = server.Serve(listener)
	}()
	healthsupport.WaitForHealthy(server)

	resp, _ := http.Get(fmt.Sprintf("http://%s/metrics", server.Addr))
	body, _ := io.ReadAll(resp.Body)
	assert.Contains(t, string(body), "TYPE any_request_duration_seconds histogram")
	_ = server.Shutdown(context.Background())
}
