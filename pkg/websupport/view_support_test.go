package websupport_test

import (
	"bytes"
	"io"
	"net"
	"net/http/httptest"
	"testing"

	"github.com/gorilla/mux"
	"github.com/hexa-org/policy-orchestrator/pkg/websupport"
	"github.com/hexa-org/policy-orchestrator/pkg/websupport/test"
	"github.com/stretchr/testify/assert"
)

func TestModelAndView(t *testing.T) {
	listener, _ := net.Listen("tcp", "localhost:0")
	websupport.Create(listener.Addr().String(), func(x *mux.Router) {}, websupport.Options{})
	writer := &httptest.ResponseRecorder{Body: new(bytes.Buffer)}

	_ = websupport.ModelAndView(
		writer,
		&websupport_test.Resources,
		"test",
		websupport.Model{Map: map[string]interface{}{"resource": "resource"}},
	)
	body, _ := io.ReadAll(writer.Body)
	assert.Contains(t, string(body), "success!")
	assert.Contains(t, string(body), "Resource")
	assert.Contains(t, string(body), "contains")
	assert.Contains(t, string(body), "nope")

	err := websupport.ModelAndView(
		&httptest.ResponseRecorder{},
		&websupport_test.Resources,
		"bad",
		websupport.Model{},
	)
	assert.Contains(t, err.Error(), "can't evaluate field Ba")
}
