package web_support_test

import (
	"bytes"
	"fmt"
	"github.com/gorilla/mux"
	"github.com/hexa-org/policy-orchestrator/pkg/web_support"
	"github.com/stretchr/testify/assert"
	"io"
	"net"
	"net/http"
	"net/http/httptest"
	"path/filepath"
	"runtime"
	"testing"
)

func TestHealth(t *testing.T) {
	listener, _ := net.Listen("tcp", "localhost:0")
	server := web_support.Create(listener.Addr().String(), func(x *mux.Router) {}, web_support.Options{})
	go web_support.Start(server, listener)

	web_support.WaitForHealthy(server)

	resp, _ := http.Get(fmt.Sprintf("http://%s/health", server.Addr))
	body, _ := io.ReadAll(resp.Body)
	assert.Equal(t, "{\"status\":\"pass\"}", string(body))

	web_support.Stop(server)
}

func TestWaitForHealth(t *testing.T) {
	listener, _ := net.Listen("tcp", "localhost:0")
	server := web_support.Create(listener.Addr().String(), func(x *mux.Router) {}, web_support.Options{})

	go web_support.Start(server, listener)
	web_support.WaitForHealthy(server)

	resp, _ := http.Get(fmt.Sprintf("http://%s/health", server.Addr))
	assert.Equal(t, http.StatusOK, resp.StatusCode)

	web_support.Stop(server)
}

func TestModelAndView(t *testing.T) {
	_, file, _, _ := runtime.Caller(0)
	resourcesDirectory := filepath.Join(file, "../../../pkg/web_support/test")
	options := web_support.Options{ResourceDirectory: resourcesDirectory}

	listener, _ := net.Listen("tcp", "localhost:0")
	web_support.Create(listener.Addr().String(), func(x *mux.Router) {}, options)
	writer := &httptest.ResponseRecorder{Body: new(bytes.Buffer)}

	_ = web_support.ModelAndView(writer, "test", web_support.Model{Map: map[string]interface{}{"resource": "resource"}})
	body, _ := io.ReadAll(writer.Body)
	assert.Contains(t, string(body), "success!")
	assert.Contains(t, string(body), "Resource")
	assert.Contains(t, string(body), "contains")
	assert.Contains(t, string(body), "nope")

	err := web_support.ModelAndView(&httptest.ResponseRecorder{}, "bad", web_support.Model{})
	assert.Contains(t, err.Error(), "can't evaluate field Ba")
}
