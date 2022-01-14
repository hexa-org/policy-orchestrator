package main

import (
	"fmt"
	"github.com/hexa-org/policy-orchestrator/pkg/web_support"
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
		web_support.Start(app, listener)
		scheduler.Start()
	}()
	web_support.WaitForHealthy(app)
	resp, _ := http.Get(fmt.Sprintf("http://%s/health", app.Addr))
	assert.Equal(t, http.StatusOK, resp.StatusCode)
	web_support.Stop(app)
	scheduler.Stop()
}

func TestConfigWithPort(t *testing.T) {
	_ = os.Setenv("PORT", "0")
	newApp()
}
