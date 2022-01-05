package web_support_test

import (
	"bytes"
	"github.com/gorilla/mux"
	"github.com/stretchr/testify/assert"
	"hexa/pkg/web_support"
	"io"
	"log"
	"net/http"
	"net/http/httptest"
	"path/filepath"
	"runtime"
	"testing"
)

func TestHealth(t *testing.T) {
	server := web_support.Create("localhost:8883", func(x *mux.Router) {}, web_support.Options{})
	go web_support.Start(server)

	web_support.WaitForHealthy(server)

	resp, err := http.Get("http://localhost:8883/health")
	if err != nil {
		log.Fatalln(err)
	}
	body, _ := io.ReadAll(resp.Body)
	assert.Equal(t, "{\"status\":\"pass\"}", string(body))

	web_support.Stop(server)
}

func TestWaitForHealth(t *testing.T) {
	server := web_support.Create("localhost:8883", func(x *mux.Router) {}, web_support.Options{})

	go web_support.Start(server)
	web_support.WaitForHealthy(server)

	resp, _ := http.Get("http://localhost:8883/health")
	assert.Equal(t, http.StatusOK, resp.StatusCode)

	web_support.Stop(server)
}

func TestModelAndView(t *testing.T) {
	_, file, _, _ := runtime.Caller(0)
	resourcesDirectory := filepath.Join(file, "../../../pkg/web_support/test")
	options := web_support.Options{ResourceDirectory: resourcesDirectory}

	web_support.Create("localhost:8883", func(x *mux.Router) {}, options)
	writer := &httptest.ResponseRecorder{Body: new(bytes.Buffer)}

	_ = web_support.ModelAndView(writer, "test", web_support.Model{})
	body, _ := io.ReadAll(writer.Body)
	assert.Contains(t, string(body), "success!")

	err := web_support.ModelAndView(&httptest.ResponseRecorder{}, "bad", web_support.Model{})
	assert.Contains(t, err.Error(), "can't evaluate field Ba")
}
