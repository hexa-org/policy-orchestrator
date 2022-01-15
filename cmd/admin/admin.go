package main

import (
	"fmt"
	"github.com/hexa-org/policy-orchestrator/pkg/admin"
	"github.com/hexa-org/policy-orchestrator/pkg/websupport"
	"log"
	"net"
	"net/http"
	"os"
	"path/filepath"
	"runtime"
)

func App(resourcesDirectory string, addr string, orchestratorUrl string, orchestratorKey string) *http.Server {
	client := admin.NewOrchestratorClient(&http.Client{}, orchestratorKey)
	handlers := admin.LoadHandlers(orchestratorUrl, client)
	options := websupport.Options{ResourceDirectory: resourcesDirectory}
	return websupport.Create(addr, handlers, options)
}

func newApp(addr string) (*http.Server, net.Listener) {
	if found := os.Getenv("PORT"); found != "" {
		host, _, _ := net.SplitHostPort(addr)
		addr = fmt.Sprintf("%v:%v", host, found)
	}
	log.Printf("Found server address %v", addr)

	orchestratorUrl := os.Getenv("ORCHESTRATOR_URL")
	orchestratorKey := os.Getenv("ORCHESTRATOR_KEY")
	_, file, _, _ := runtime.Caller(0)
	resourcesDirectory := filepath.Join(file, "../../../pkg/admin/resources")
	listener, _ := net.Listen("tcp", addr)
	return App(resourcesDirectory, listener.Addr().String(), orchestratorUrl, orchestratorKey), listener
}

func main() {
	websupport.Start(newApp("0.0.0.0:8884"))
}
