package main

import (
	"errors"
	"fmt"
	"github.com/gorilla/mux"
	"github.com/hexa-org/policy-orchestrator/pkg/healthsupport"
	"github.com/hexa-org/policy-orchestrator/pkg/websupport"
	"github.com/stretchr/testify/assert"
	"io"
	"net"
	"net/http"
	"os"
	"path/filepath"
	"runtime"
	"testing"
)

type ErroneousWorker struct {
	Err error
}

func (m *ErroneousWorker) Run(i interface{}) error {
	m.Err = errors.New("oops")
	return m.Err
}

func TestProxy(t *testing.T) {

	listener, _ := net.Listen("tcp", "localhost:0")
	app := aSmallTestApp(listener)

	remoteUrl := fmt.Sprintf("http://%s", listener.Addr().String())

	_, file, _, _ := runtime.Caller(0)
	resourcesDirectory := filepath.Join(file, "../resources")
	proxyListener, _ := net.Listen("tcp", "localhost:0")
	proxyApp := App(remoteUrl, proxyListener.Addr().String(), resourcesDirectory, IdentityWorker{})
	go func() {
		websupport.Start(proxyApp, proxyListener)
	}()
	healthsupport.WaitForHealthy(proxyApp)

	dashboard, _ := http.Get(fmt.Sprintf("http://%s/_proxy", proxyApp.Addr))
	dashboardBody, _ := io.ReadAll(dashboard.Body)
	assert.Contains(t, string(dashboardBody), "Welcome to Hexa Proxy")

	proxy, _ := http.Get(fmt.Sprintf("http://%s/", proxyApp.Addr))
	proxyBody, _ := io.ReadAll(proxy.Body)
	assert.Equal(t, "success", string(proxyBody))

	proxyPlus, _ := http.Get(fmt.Sprintf("http://%s/plus", proxyApp.Addr))
	proxyPlusBody, _ := io.ReadAll(proxyPlus.Body)
	assert.Equal(t, "plus", string(proxyPlusBody))

	websupport.Stop(proxyApp)
	websupport.Stop(app)
}

func TestProxy_WithError(t *testing.T) {

	worker := &ErroneousWorker{}

	listener, _ := net.Listen("tcp", "localhost:0")
	app := aSmallTestApp(listener)

	remoteUrl := fmt.Sprintf("http://%s", listener.Addr().String())

	_, file, _, _ := runtime.Caller(0)
	resourcesDirectory := filepath.Join(file, "../resources")
	proxyListener, _ := net.Listen("tcp", "localhost:0")
	proxyApp := App(remoteUrl, proxyListener.Addr().String(), resourcesDirectory, worker)
	go func() {
		websupport.Start(proxyApp, proxyListener)
	}()
	healthsupport.WaitForHealthy(proxyApp)

	_, _ = http.Get(fmt.Sprintf("http://%s/", proxyApp.Addr))
	assert.Equal(t, "oops", worker.Err.Error())

	websupport.Stop(proxyApp)
	websupport.Stop(app)
}

func TestIdentityWorker_Run(t *testing.T) {
	err := IdentityWorker{}.Run([]IdentityInfo{{"someWork"}, {"someOtherWork"}})
	assert.NoError(t, err)
}

func TestConfigWithPort(t *testing.T) {
	_ = os.Setenv("PORT", "0")
	_ = os.Setenv("REMOTER_URL", "http://localhost:8886")
	newApp("localhost:0")
}

func aSmallTestApp(listener net.Listener) *http.Server {
	app := websupport.Create(listener.Addr().String(), func(x *mux.Router) {
		x.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
			_, _ = w.Write([]byte("success"))
		})
		x.HandleFunc("/plus", func(w http.ResponseWriter, r *http.Request) {
			_, _ = w.Write([]byte("plus"))
		})
	}, websupport.Options{})
	go func() {
		websupport.Start(app, listener)
	}()
	healthsupport.WaitForHealthy(app)
	return app
}
