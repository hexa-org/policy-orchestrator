package main

import (
	"fmt"
	"github.com/hexa-org/policy-orchestrator/pkg/healthsupport"
	"github.com/hexa-org/policy-orchestrator/pkg/websupport"
	"github.com/stretchr/testify/assert"
	"net"
	"net/http"
	"os"
	"path/filepath"
	"runtime"
	"testing"
)

func TestNewApp(t *testing.T) {
	_ = os.Setenv("PORT", "0")
	newApp("localhost:0")
}

func TestApp(t *testing.T) {
	_, file, _, _ := runtime.Caller(0)
	resourcesDirectory := filepath.Join(file, "../../../cmd/demo/resources")
	listener, _ := net.Listen("tcp", "localhost:0")
	app := App(listener.Addr().String(), resourcesDirectory)
	go func() {
		websupport.Start(app, listener)
	}()
	healthsupport.WaitForHealthy(app)

	response, _ := http.Get(fmt.Sprintf("http://%s/health", app.Addr))
	assert.Equal(t, http.StatusOK, response.StatusCode)
	websupport.Stop(app)
}

func TestDownload(t *testing.T) {
	_, file, _, _ := runtime.Caller(0)
	resourcesDirectory := filepath.Join(file, "../../../cmd/demo/resources")
	listener, _ := net.Listen("tcp", "localhost:0")
	app := App(listener.Addr().String(), resourcesDirectory)
	go func() {
		websupport.Start(app, listener)
	}()
	healthsupport.WaitForHealthy(app)

	response, _ := http.Get(fmt.Sprintf("http://%s/bundles/bundle.tar.gz", app.Addr))
	assert.Equal(t, http.StatusOK, response.StatusCode)
	websupport.Stop(app)
}
