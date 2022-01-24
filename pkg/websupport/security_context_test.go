package websupport_test

import (
	"bytes"
	"github.com/gorilla/mux"
	"github.com/hexa-org/policy-orchestrator/pkg/websupport"
	"github.com/stretchr/testify/assert"
	"io"
	"net"
	"net/http"
	"net/http/httptest"
	"path/filepath"
	"runtime"
	"strings"
	"testing"
)

func TestSecurityContextMiddleware_withEmail(t *testing.T) {
	model := websupport.Model{Map: map[string]interface{}{"resource": "resource"}}
	request := httptest.NewRequest("GET", "/", strings.NewReader(""))
	request.Header["X-Goog-Authenticated-User-Email"] = []string{"anEmail@google.com"}

	writer := makeRequest(request, model)

	body, _ := io.ReadAll(writer.Body)
	assert.Contains(t, string(body), "anEmail@google.com")
}

func TestSecurityContextMiddleware_withoutEmail(t *testing.T) {
	model := websupport.Model{Map: map[string]interface{}{"resource": "resource"}}
	request := httptest.NewRequest("GET", "/", strings.NewReader(""))

	writer := makeRequest(request, model)

	body, _ := io.ReadAll(writer.Body)
	assert.NotContains(t, string(body), "anEmail@google.com")
}

func makeRequest(request *http.Request, model websupport.Model) *httptest.ResponseRecorder {
	_, file, _, _ := runtime.Caller(0)
	resourcesDirectory := filepath.Join(file, "../../../pkg/websupport/test")
	options := websupport.Options{ResourceDirectory: resourcesDirectory}

	listener, _ := net.Listen("tcp", "localhost:0")
	websupport.Create(listener.Addr().String(), func(x *mux.Router) {}, options)

	writer := &httptest.ResponseRecorder{Body: new(bytes.Buffer)}
	_ = websupport.ModelAndView(websupport.SecurityContextViewSupport(writer, request, "security", model))
	return writer
}
