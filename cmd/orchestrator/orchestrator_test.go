package main

import (
	"github.com/hexa-org/policy-orchestrator/pkg/web_support"
	"github.com/stretchr/testify/assert"
	"net/http"
	"os"
	"testing"
)

func TestApp(t *testing.T) {
	app, scheduler := App("aKey", "localhost:8883", "localhost:8883",
		"postgres://orchestrator:orchestrator@localhost:5432/orchestrator_test?sslmode=disable")
	go func() {
		web_support.Start(app)
		scheduler.Start()
	}()
	web_support.WaitForHealthy(app)
	resp, _ := http.Get("http://localhost:8883/health")
	assert.Equal(t, http.StatusOK, resp.StatusCode)
	web_support.Stop(app)
	scheduler.Stop()
}

func TestConfig(t *testing.T) {
	newApp()
}

func TestConfigWithPort(t *testing.T) {
	_ = os.Setenv("PORT", "8883")
	newApp()
}
