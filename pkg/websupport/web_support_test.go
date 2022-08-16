package websupport_test

import (
	"crypto/tls"
	"crypto/x509"
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
	assert.Contains(t, string(body), "TYPE any_request_duration_seconds histogram")

	websupport.Stop(server)
}

func _TestServerWithTLS(t *testing.T) {
	_, file, _, _ := runtime.Caller(0)

	listener, _ := net.Listen("tcp", "localhost:0")
	server := websupport.Create(listener.Addr().String(), func(x *mux.Router) {}, websupport.Options{})
	configureWithTransportLayerSecurity(file, server)
	go websupport.StartWithTLS(server, listener)

	caCert := must(os.ReadFile(filepath.Join(file, "../test/ca-cert.pem")))
	clientCert, _ := tls.X509KeyPair(
		must(os.ReadFile(filepath.Join(file, "../test/client-cert.pem"))),
		must(os.ReadFile(filepath.Join(file, "../test/client-key.pem"))),
	)
	caCertPool := x509.NewCertPool()
	caCertPool.AppendCertsFromPEM(caCert)

	client := &http.Client{
		Transport: &http.Transport{
			TLSClientConfig: &tls.Config{
				Certificates: []tls.Certificate{clientCert},
				RootCAs:      caCertPool,
			},
		},
	}
	healthsupport.WaitForHealthyWithClient(server, client, fmt.Sprintf("https://%s/health", server.Addr))

	websupport.Stop(server)
}

func TestPaths(t *testing.T) {
	listener, _ := net.Listen("tcp", "localhost:0")
	server := websupport.Create(listener.Addr().String(), func(x *mux.Router) {}, websupport.Options{})

	assert.Equal(t, 2, len(websupport.Paths(server.Handler.(*mux.Router))))
}

/// supporting functions

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
