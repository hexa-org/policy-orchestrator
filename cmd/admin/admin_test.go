package main

import (
	"github.com/hexa-org/policy-orchestrator/pkg/web_support"
	"github.com/stretchr/testify/assert"
	"net/http"
	"os"
	"path/filepath"
	"runtime"
	"testing"
)

func TestApp(t *testing.T) {
	_, file, _, _ := runtime.Caller(0)
	resourcesDirectory := filepath.Join(file, "../../../pkg/admin/resources")
	app := App(resourcesDirectory, "localhost:8883", "localhost:8883", "aKey")
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

func TestConfigWithPort(t *testing.T) {
	_ = os.Setenv("PORT", "8883")
	newApp()
}
