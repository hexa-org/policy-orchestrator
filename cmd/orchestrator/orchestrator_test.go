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

	"github.com/hexa-org/policy-orchestrator/pkg/healthsupport"
	"github.com/hexa-org/policy-orchestrator/pkg/websupport"
	"github.com/stretchr/testify/assert"
)

func TestApp(t *testing.T) {
	listener, _ := net.Listen("tcp", "localhost:0")
	app, scheduler := App("aKey", listener.Addr().String(), listener.Addr().String(), "postgres://orchestrator:orchestrator@localhost:5432/orchestrator_test?sslmode=disable")
	go func() {
		websupport.Start(app, listener)
		scheduler.Start()
	}()
	healthsupport.WaitForHealthy(app)

	get, _ := http.Get(fmt.Sprintf("http://%s/health", app.Addr))
	body, _ := io.ReadAll(get.Body)
	assert.Equal(t, "[{\"name\":\"server\",\"pass\":\"true\"},{\"name\":\"database\",\"pass\":\"true\"}]", string(body))

	websupport.Stop(app)
	scheduler.Stop()
}

func TestAppWithTransportLayerSecurity(t *testing.T) {
	listener, _ := net.Listen("tcp", "localhost:0")
	app, scheduler := App("aKey", listener.Addr().String(), listener.Addr().String(), "postgres://orchestrator:orchestrator@localhost:5432/orchestrator_test?sslmode=disable")
	_, file, _, _ := runtime.Caller(0)
	configureWithTransportLayerSecurity(file, app)
	go func() {
		websupport.Start(app, listener)
		scheduler.Start()
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
	scheduler.Stop()
}

func TestConfigWithPort(t *testing.T) {
	t.Setenv("PORT", "0")
	t.Setenv("HOST", "localhost")
	newApp("localhost:0")
}

func TestConfigWithTransportLayerSecurity(t *testing.T) {
	_, file, _, _ := runtime.Caller(0)
	t.Setenv("SERVER_CERT", filepath.Join(file, "../test/server-cert.pem"))
	t.Setenv("SERVER_KEY", filepath.Join(file, "../test/server-key.pem"))
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
