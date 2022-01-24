package websupport_test

import (
	"bytes"
	"github.com/gorilla/mux"
	"github.com/hexa-org/policy-orchestrator/pkg/websupport"
	"github.com/stretchr/testify/assert"
	"io"
	"net"
	"net/http/httptest"
	"path/filepath"
	"runtime"
	"testing"
)

func TestModelAndView(t *testing.T) {
	_, file, _, _ := runtime.Caller(0)
	resourcesDirectory := filepath.Join(file, "../../../pkg/websupport/test")
	options := websupport.Options{ResourceDirectory: resourcesDirectory}

	listener, _ := net.Listen("tcp", "localhost:0")
	websupport.Create(listener.Addr().String(), func(x *mux.Router) {}, options)
	writer := &httptest.ResponseRecorder{Body: new(bytes.Buffer)}

	_ = websupport.ModelAndView(writer, "test", websupport.Model{Map: map[string]interface{}{"resource": "resource"}})
	body, _ := io.ReadAll(writer.Body)
	assert.Contains(t, string(body), "success!")
	assert.Contains(t, string(body), "Resource")
	assert.Contains(t, string(body), "contains")
	assert.Contains(t, string(body), "nope")

	err := websupport.ModelAndView(&httptest.ResponseRecorder{}, "bad", websupport.Model{})
	assert.Contains(t, err.Error(), "can't evaluate field Ba")
}
