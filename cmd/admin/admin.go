package main

import (
	"fmt"
	"github.com/hexa-org/policy-orchestrator/pkg/admin"
	"github.com/hexa-org/policy-orchestrator/pkg/web_support"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"runtime"
)

func App(resourcesDirectory string, addr string, orchestratorUrl string, orchestratorKey string) *http.Server {
	client := admin.NewOrchestratorClient(&http.Client{}, orchestratorKey)
	handlers := admin.LoadHandlers(orchestratorUrl, client)
	options := web_support.Options{ResourceDirectory: resourcesDirectory}
	return web_support.Create(addr, handlers, options)
}

func newApp() *http.Server {
	addr := "0.0.0.0:8884"
	if found := os.Getenv("PORT"); found != "" {
		addr = fmt.Sprintf("0.0.0.0:%v", found)
	}
	log.Printf("Found server address %v", addr)

	orchestratorUrl := os.Getenv("ORCHESTRATOR_URL")
	orchestratorKey := os.Getenv("ORCHESTRATOR_KEY")
	_, file, _, _ := runtime.Caller(0)
	resourcesDirectory := filepath.Join(file, "../../../pkg/admin/resources")

	server := App(resourcesDirectory, addr, orchestratorUrl, orchestratorKey)
	return server
}

func main() {
	web_support.Start(newApp())
}
