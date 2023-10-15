package websupport_test

import (
	"crypto/tls"
	"fmt"
	"io"
	"net"
	"net/http"
	"os"
	"path/filepath"
	"runtime"
	"testing"

	"github.com/gorilla/mux"
	"github.com/hexa-org/policy-orchestrator/demo/pkg/healthsupport"
	"github.com/hexa-org/policy-orchestrator/demo/pkg/websupport"
	"github.com/stretchr/testify/assert"
)

func TestServer(t *testing.T) {
	listener, _ := net.Listen("tcp", "localhost:0")
	server := websupport.Create(listener.Addr().String(), func(x *mux.Router) {}, websupport.Options{})
	go websupport.Start(server, listener)

	healthsupport.WaitForHealthy(server)

	resp, _ := http.Get(fmt.Sprintf("http://%s/health", server.Addr))
	body, _ := io.ReadAll(resp.Body)
	assert.Equal(t, "[{\"name\":\"noop\",\"pass\":\"true\"}]", string(body))

	resp, _ = http.Get(fmt.Sprintf("http://%s/metrics", server.Addr))
	body, _ = io.ReadAll(resp.Body)
	assert.Contains(t, string(body), "{}")

	websupport.Stop(server)
}

func TestWithTransportLayerSecurity(t *testing.T) {
	_, file, _, _ := runtime.Caller(0)
	certPath := filepath.Join(file, "../test/certs/server-cert.pem")
	keyPath := filepath.Join(file, "../test/certs/server-key.pem")

	server := &http.Server{}

	websupport.WithTransportLayerSecurity(certPath, keyPath, server)

	assert.NotNil(t, server.TLSConfig)
}

func TestPaths(t *testing.T) {
	listener, _ := net.Listen("tcp", "localhost:0")
	server := websupport.Create(listener.Addr().String(), func(x *mux.Router) {}, websupport.Options{})

	assert.Equal(t, 2, len(websupport.Paths(server.Handler.(*mux.Router))))
}

// / supporting functions

func configureWithTransportLayerSecurity(file string, server *http.Server) {
	cert, _ := tls.X509KeyPair(
		must(os.ReadFile(filepath.Join(file, "../test/server-cert.pem"))),
		must(os.ReadFile(filepath.Join(file, "../test/server-key.pem"))),
	)
	server.TLSConfig = &tls.Config{
		Certificates: []tls.Certificate{cert},
	}
}

func must(file []byte, err error) []byte {
	if err != nil {
		panic("unable to read file.")
	}
	return file
}
