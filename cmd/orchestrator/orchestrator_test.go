package main

import (
	"github.com/stretchr/testify/assert"
	"hexa/pkg/web_support"
	"net/http"
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
