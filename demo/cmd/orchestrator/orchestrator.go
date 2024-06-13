package main

import (
	"fmt"
	"net"
	"net/http"
	"os"

	"github.com/hexa-org/policy-opa/pkg/keysupport"
	"github.com/hexa-org/policy-orchestrator/demo/internal/orchestrator"
	"github.com/hexa-org/policy-orchestrator/demo/pkg/dataConfigGateway"

	log "golang.org/x/exp/slog"

	"github.com/hexa-org/policy-orchestrator/demo/pkg/healthsupport"
	"github.com/hexa-org/policy-orchestrator/demo/pkg/websupport"
)

type ServerHealthCheck struct {
}

func (s ServerHealthCheck) Name() string {
	return "server"
}

func (s ServerHealthCheck) Check() bool {
	return true
}

func App(key string, addr string, hostPort string) *http.Server {

	config, err := dataConfigGateway.NewIntegrationConfigData()
	if err != nil {
		panic(err)
	}

	handlers := orchestrator.LoadHandlers(config, nil)
	return websupport.Create(addr, handlers, websupport.Options{
		HealthChecks: []healthsupport.HealthCheck{
			ServerHealthCheck{},
		},
	})
}

func newApp(addr string) (*http.Server, net.Listener) {
	if found := os.Getenv("PORT"); found != "" {
		host, _, _ := net.SplitHostPort(addr)
		addr = fmt.Sprintf("%v:%v", host, found)
	}
	log.Info("Orchestrator Start", "Found server address", addr)

	if found := os.Getenv("HOST"); found != "" {
		_, port, _ := net.SplitHostPort(addr)
		addr = fmt.Sprintf("%v:%v", found, port)
	}

	log.Info("Orchestrator Start", "Found server host", addr)

	key := os.Getenv("ORCHESTRATOR_KEY")
	hostPort := os.Getenv("ORCHESTRATOR_HOSTPORT")
	listener, _ := net.Listen("tcp", addr)
	app := App(key, listener.Addr().String(), hostPort)

	if websupport.IsTlsEnabled() {
		keyConfig := keysupport.GetKeyConfig()
		err := keyConfig.InitializeKeys()
		if err != nil {
			log.Error("Error initializing keys: " + err.Error())
			panic(err)
		}

		websupport.WithTransportLayerSecurity(keyConfig.ServerCertPath, keyConfig.ServerKeyPath, app)
	}

	return app, listener
}

func main() {
	app, listener := newApp("0.0.0.0:8885")

	websupport.Start(app, listener)
}
