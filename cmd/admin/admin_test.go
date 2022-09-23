package main

import (
	"fmt"
	"net"
	"net/http"
	"testing"

	"github.com/hexa-org/policy-orchestrator/pkg/healthsupport"
	"github.com/hexa-org/policy-orchestrator/pkg/websupport"
	"github.com/stretchr/testify/assert"
)

func TestApp(t *testing.T) {
	listener, _ := net.Listen("tcp", "localhost:0")
	app := App(listener.Addr().String(), "http://localhost:8885/", "aKey")
	go func() {
		websupport.Start(app, listener)
	}()
	healthsupport.WaitForHealthy(app)
	resp, _ := http.Get(fmt.Sprintf("http://%s/health", app.Addr))
	assert.Equal(t, http.StatusOK, resp.StatusCode)
	websupport.Stop(app)
}

func TestConfigWithPort(t *testing.T) {
	t.Setenv("PORT", "0")
	newApp("localhost:0")

}
