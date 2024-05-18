package main

import (
	"errors"
	"fmt"
	"log"
	"net"
	"net/http"
	"os"

	"github.com/hexa-org/policy-orchestrator/demo/internal/admin"

	"github.com/hexa-org/policy-orchestrator/demo/pkg/websupport"
)

func App(addr string, orchestratorUrl string, orchestratorKey string) *http.Server {
	client := admin.NewOrchestratorClient(&http.Client{}, orchestratorUrl, orchestratorKey)
	handlers := admin.LoadHandlers(orchestratorUrl, client)
	return websupport.Create(addr, handlers, websupport.Options{})
}

func newApp(addr string) (*http.Server, net.Listener) {
	if found := os.Getenv("PORT"); found != "" {
		host, _, _ := net.SplitHostPort(addr)
		addr = fmt.Sprintf("%v:%v", host, found)
	}
	log.Printf("Found server address %v", addr)

	orchestratorUrl := os.Getenv("ORCHESTRATOR_URL")
	orchestratorKey := os.Getenv("ORCHESTRATOR_KEY")
	if orchestratorKey == "" || orchestratorUrl == "" {
		err := errors.New("Fatal error: ORCHESTRATOR_URL or ORCHESTRATOR_KEY environment variable not set")
		log.Println(err.Error())
		panic(err)
	}
	listener, _ := net.Listen("tcp", addr)
	return App(listener.Addr().String(), orchestratorUrl, orchestratorKey), listener
}

func main() {
	websupport.Start(newApp("0.0.0.0:8884"))
}
