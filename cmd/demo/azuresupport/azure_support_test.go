package azuresupport_test

import (
	"fmt"
	"io"
	"net"
	"net/http"
	"testing"

	"github.com/gorilla/mux"
	"github.com/gorilla/sessions"
	"github.com/hexa-org/policy-orchestrator/cmd/demo/azuresupport"
	"github.com/hexa-org/policy-orchestrator/pkg/healthsupport"
	"github.com/hexa-org/policy-orchestrator/pkg/websupport"
	"github.com/stretchr/testify/assert"
)

func TestAzureSecurityOptions(t *testing.T) {
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
	}, websupport.Options{})
	router := server.Handler.(*mux.Router)
	router.Use(azuresupport.NewAzureSupport(session).Middleware)

	go websupport.Start(server, listener)
	healthsupport.WaitForHealthy(server)
	defer websupport.Stop(server)

	request, _ := http.NewRequest("GET", fmt.Sprintf("http://%s/", server.Addr), nil)
	request.Header["X-Ms-Client-Principal-Name"] = []string{"example@microsoft.com"}
	response, _ := (&http.Client{}).Do(request)

	body, _ := io.ReadAll(response.Body)
	assert.Contains(t, string(body), "example@microsoft.com")
}
