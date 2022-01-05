package main

import (
	"github.com/stretchr/testify/assert"
	"hexa/pkg/web_support"
	"net/http"
	"testing"
)

func TestApp(t *testing.T) {
	app := App("aKey", "localhost:8883", "localhost:8883")
	go func() {
		web_support.Start(app)
	}()
	web_support.WaitForHealthy(app)
	resp, _ := http.Get("http://localhost:8883/health")
	assert.Equal(t, http.StatusOK, resp.StatusCode)
	web_support.Stop(app)
}

func TestConfig(t *testing.T) {
	newApp()
}
