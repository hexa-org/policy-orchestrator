package main

import (
	"fmt"
	"net"
	"net/http"
	"os"

	"github.com/hexa-org/policy-mapper/pkg/keysupport"
	"github.com/hexa-org/policy-mapper/pkg/sessionSupport"
	"github.com/hexa-org/policy-orchestrator/demo/pkg/hexaConstants"
	log "golang.org/x/exp/slog"

	"github.com/hexa-org/policy-orchestrator/demo/internal/admin"

	"github.com/hexa-org/policy-mapper/pkg/websupport"
)

func App(addr string, orchestratorUrl string) *http.Server {

	client := admin.NewOrchestratorClient(nil, orchestratorUrl)

	sessionHandler := sessionSupport.NewSessionManager()

	handlers := admin.LoadHandlers(orchestratorUrl, client, sessionHandler)
	return websupport.Create(addr, handlers, websupport.Options{})
}

func newApp(addr string) (*http.Server, net.Listener) {
	if found := os.Getenv("PORT"); found != "" {
		host, _, _ := net.SplitHostPort(addr)
		addr = fmt.Sprintf("%v:%v", host, found)
	}
	log.Debug("Found server address %v", addr)

	orchestratorUrl := os.Getenv("ORCHESTRATOR_URL")

	listener, _ := net.Listen("tcp", addr)
	server := App(listener.Addr().String(), orchestratorUrl)

	if websupport.IsTlsEnabled() {
		keyConfig := keysupport.GetKeyConfig()
		err := keyConfig.InitializeKeys()
		if err != nil {
			log.Error("Error initializing keys: " + err.Error())
			panic(err)
		}

		websupport.WithTransportLayerSecurity(keyConfig.ServerCertPath, keyConfig.ServerKeyPath, server)
	}
	return server, listener
}

func main() {
	log.Info("Hexa Orchestrator Admin UI Server starting...", "version", hexaConstants.HexaOrchestratorVersion)
	websupport.Start(newApp("0.0.0.0:8884"))
}
