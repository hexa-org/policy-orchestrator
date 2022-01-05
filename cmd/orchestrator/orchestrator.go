package main

import (
	"fmt"
	"hexa/pkg/hawk_support"
	"hexa/pkg/orchestrator"
	"hexa/pkg/web_support"
	"log"
	"net/http"
	"os"
)

func App(key string, addr string, hostPort string) *http.Server {
	store := hawk_support.NewCredentialStore(key)
	handlers := orchestrator.LoadHandlers(store, hostPort)
	return web_support.Create(addr, handlers, web_support.Options{})
}

func newApp() *http.Server {
	addr := "0.0.0.0:8885"
	if found := os.Getenv("PORT"); found != "" {
		addr = fmt.Sprintf("0.0.0.0:%v", found)
	}
	log.Printf("Found server address %v", addr)

	key := os.Getenv("ORCHESTRATOR_KEY")
	hostPort := os.Getenv("ORCHESTRATOR_HOSTPORT")
	app := App(key, addr, hostPort)
	return app
}

func main() {
	web_support.Start(newApp())
}
