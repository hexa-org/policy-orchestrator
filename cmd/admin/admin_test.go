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

func TestApp(t *testing.T) {
	_, file, _, _ := runtime.Caller(0)
	resourcesDirectory := filepath.Join(file, "../../../pkg/admin/resources")
	listener, _ := net.Listen("tcp", "localhost:0")
	app := App(resourcesDirectory, listener.Addr().String(), "http://localhost:8885/", "aKey")
	go func() {
		websupport.Start(app, listener)
	}()
	healthsupport.WaitForHealthy(app)
	resp, _ := http.Get(fmt.Sprintf("http://%s/health", app.Addr))
	assert.Equal(t, http.StatusOK, resp.StatusCode)
	websupport.Stop(app)
}

func TestConfigWithPort(t *testing.T) {
	_ = os.Setenv("PORT", "0")
	newApp("localhost:0")
}
