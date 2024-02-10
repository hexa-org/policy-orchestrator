package metricssupport_test

import (
	"context"
	"fmt"
	"log"

	"github.com/hexa-org/policy-orchestrator/demo/pkg/metricssupport"

	"github.com/gorilla/mux"

	"io"
	"net"
	"net/http"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestMetrics(t *testing.T) {
	listener, _ := net.Listen("tcp", "localhost:0")
	router := mux.NewRouter()
	router.Use(metricssupport.MetricsMiddleware)
	router.Path("/metrics").Handler(metricssupport.MetricsHandler())

	server := &http.Server{
		Addr:    listener.Addr().String(),
		Handler: router,
	}
	go func() {
		_ = server.Serve(listener)
	}()
	waitForHealthy(server)

	resp, _ := http.Get(fmt.Sprintf("http://%s/metrics", server.Addr))
	body, _ := io.ReadAll(resp.Body)
	assert.Contains(t, string(body), "{}")
	_ = server.Shutdown(context.Background())
}

func waitForHealthy(server *http.Server) {
	var isLive bool
	for !isLive {
		resp, err := http.Get(fmt.Sprintf("http://%s/metrics", server.Addr))
		if err == nil && resp.StatusCode == http.StatusOK {
			log.Println("Server is healthy.", server.Addr)
			isLive = true
		}
	}
}
