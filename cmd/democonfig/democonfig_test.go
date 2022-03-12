package main

import (
	"bytes"
	"fmt"
	"github.com/hexa-org/policy-orchestrator/pkg/compressionsupport"
	"github.com/hexa-org/policy-orchestrator/pkg/healthsupport"
	"github.com/hexa-org/policy-orchestrator/pkg/websupport"
	"github.com/stretchr/testify/assert"
	"mime/multipart"
	"net"
	"net/http"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"testing"
)

func TestNewApp(t *testing.T) {
	_ = os.Setenv("PORT", "0")
	newApp("localhost:0")
}

func setup() *http.Server {
	_, file, _, _ := runtime.Caller(0)
	resourcesDirectory := filepath.Join(file, "../resources")
	listener, _ := net.Listen("tcp", "localhost:0")
	app := App(listener.Addr().String(), resourcesDirectory)
	go func() {
		websupport.Start(app, listener)
	}()
	healthsupport.WaitForHealthy(app)
	return app
}

func TestApp(t *testing.T) {
	app := setup()
	response, _ := http.Get(fmt.Sprintf("http://%s/health", app.Addr))
	assert.Equal(t, http.StatusOK, response.StatusCode)
	websupport.Stop(app)
}

func TestDownload(t *testing.T) {
	app := setup()
	response, _ := http.Get(fmt.Sprintf("http://%s/bundles/bundle.tar.gz", app.Addr))
	assert.Equal(t, http.StatusOK, response.StatusCode)
	websupport.Stop(app)
}

func TestUpload(t *testing.T) {
	app := setup()

	_, file, _, _ := runtime.Caller(0)
	bundleDir := filepath.Join(file, "../resources/bundles")
	tar, _ := compressionsupport.TarFromPath(bundleDir)
	var buffer bytes.Buffer
	_ = compressionsupport.Gzip(&buffer, tar)

	buf := new(bytes.Buffer)
	writer := multipart.NewWriter(buf)
	formFile,_ := writer.CreateFormFile("bundle", "bundle.tar.gz")
	_, _ = formFile.Write(buffer.Bytes())
	_ = writer.Close()

	contentType := writer.FormDataContentType()
	response, _ := http.Post(fmt.Sprintf("http://%s/bundles", app.Addr), contentType, buf)
	assert.Equal(t, http.StatusCreated, response.StatusCode)

	glob, _ := filepath.Glob(fmt.Sprintf("%s/*", bundleDir))
	for _, item := range glob {
		if strings.Contains(item,".bundle") {
			_ = os.RemoveAll(item)
		}
	}
	websupport.Stop(app)
}
