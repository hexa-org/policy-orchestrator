package main

import (
	"fmt"
	"net"
	"net/http"
	"testing"

	"github.com/hexa-org/policy-orchestrator/demo/pkg/healthsupport"
	"github.com/hexa-org/policy-orchestrator/demo/pkg/websupport"
	"github.com/stretchr/testify/assert"
)

func TestApp(t *testing.T) {
	listener, _ := net.Listen("tcp", "localhost:0")
	app := App(listener.Addr().String(), "http://localhost:8885/", "aKey")
	go func() {
		websupport.Start(app, listener)
	}()
	healthsupport.WaitForHealthy(app)
	resp, err := http.Get(fmt.Sprintf("http://%s/health", app.Addr))
	assert.NoError(t, err)
	assert.Equal(t, http.StatusOK, resp.StatusCode)
	websupport.Stop(app)
}

func TestConfigWithPort(t *testing.T) {
	t.Setenv("PORT", "0")
	assert.NotPanics(t, func() {
		newApp("localhost:0")
	})
}
