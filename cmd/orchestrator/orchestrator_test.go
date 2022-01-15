package main

import (
	"fmt"
	"github.com/hexa-org/policy-orchestrator/pkg/websupport"
	"github.com/stretchr/testify/assert"
	"net"
	"net/http"
	"os"
	"testing"
)

func TestApp(t *testing.T) {
	listener, _ := net.Listen("tcp", "localhost:0")
	app, scheduler := App("aKey", listener.Addr().String(), listener.Addr().String(), "postgres://orchestrator:orchestrator@localhost:5432/orchestrator_test?sslmode=disable")
	go func() {
		websupport.Start(app, listener)
		scheduler.Start()
	}()
	websupport.WaitForHealthy(app)
	resp, _ := http.Get(fmt.Sprintf("http://%s/health", app.Addr))
	assert.Equal(t, http.StatusOK, resp.StatusCode)
	websupport.Stop(app)
	scheduler.Stop()
}

func TestConfigWithPort(t *testing.T) {
	_ = os.Setenv("PORT", "0")
	newApp("localhost:0")
}
