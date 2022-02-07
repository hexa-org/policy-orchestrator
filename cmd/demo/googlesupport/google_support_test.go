package googlesupport_test

import (
	"fmt"
	"github.com/gorilla/mux"
	"github.com/gorilla/sessions"
	"github.com/hexa-org/policy-orchestrator/cmd/demo/googlesupport"
	"github.com/hexa-org/policy-orchestrator/pkg/healthsupport"
	"github.com/hexa-org/policy-orchestrator/pkg/websupport"
	"github.com/stretchr/testify/assert"
	"io"
	"net"
	"net/http"
	"path/filepath"
	"runtime"
	"testing"
)

func TestGoogleSecurityOptions(t *testing.T) {
	_, file, _, _ := runtime.Caller(0)
	resourcesDirectory := filepath.Join(file, "../../../demo/test")
	options := websupport.Options{ResourceDirectory: resourcesDirectory}

	var session = sessions.NewCookieStore([]byte("super_secret"))
	listener, _ := net.Listen("tcp", "localhost:0")
	server := websupport.Create(listener.Addr().String(), func(router *mux.Router) {
		router.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
			s, err := session.Get(r, "session")
			if err != nil {
				return
			}
			principal := s.Values["principal"].([]string)
			bytes := []byte(principal[0])
			_, _ = w.Write(bytes)
		})
	}, options)
	router := server.Handler.(*mux.Router)
	router.Use(googlesupport.NewGoogleSupport(session).Middleware)

	go websupport.Start(server, listener)
	healthsupport.WaitForHealthy(server)
	defer websupport.Stop(server)

	request, _ := http.NewRequest("GET", fmt.Sprintf("http://%s/", server.Addr), nil)
	request.Header["X-Goog-Authenticated-User-Email"] = []string{"anEmail@google.com"}
	response, _ := (&http.Client{}).Do(request)

	body, _ := io.ReadAll(response.Body)
	assert.Contains(t, string(body), "anEmail@google.com")
}
