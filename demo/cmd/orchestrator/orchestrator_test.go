package main

import (
	"crypto/tls"
	"crypto/x509"
	"fmt"
	"io"
	"net"
	"net/http"
	"os"
	"path/filepath"
	"runtime"
	"testing"

	"github.com/hexa-org/policy-opa/pkg/keysupport"
	"github.com/hexa-org/policy-orchestrator/demo/pkg/healthsupport"
	"github.com/hexa-org/policy-orchestrator/demo/pkg/websupport"
	"github.com/stretchr/testify/assert"
)

func TestApp(t *testing.T) {
	listener, _ := net.Listen("tcp", "localhost:0")
	app := App("aKey", listener.Addr().String(), listener.Addr().String())

	go func() {
		websupport.Start(app, listener)

	}()
	healthsupport.WaitForHealthy(app)

	get, _ := http.Get(fmt.Sprintf("http://%s/health", app.Addr))
	body, _ := io.ReadAll(get.Body)
	assert.Equal(t, "[{\"name\":\"server\",\"pass\":\"true\"}]", string(body))

	websupport.Stop(app)

}

func TestAppWithTransportLayerSecurity(t *testing.T) {
	listener, _ := net.Listen("tcp", "localhost:0")
	app := App("aKey", listener.Addr().String(), listener.Addr().String())
	_, file, _, _ := runtime.Caller(0)
	configureWithTransportLayerSecurity(file, app)
	go func() {
		websupport.Start(app, listener)

	}()

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
	healthsupport.WaitForHealthyWithClient(app, client, fmt.Sprintf("https://%s/health", app.Addr))

	websupport.Stop(app)
}

func TestConfigWithTransportLayerSecurity(t *testing.T) {
	_, file, _, _ := runtime.Caller(0)
	t.Setenv(keysupport.EnvServerCert, filepath.Join(file, "../test/server-cert.pem"))
	t.Setenv(keysupport.EnvServerKey, filepath.Join(file, "../test/server-key.pem"))
	t.Setenv(keysupport.EnvCertDirectory, filepath.Join(file, "../test"))
	newApp("localhost:0")
}

// supporting functions

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
