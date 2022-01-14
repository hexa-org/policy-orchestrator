package main

import (
	"fmt"
	"github.com/hexa-org/policy-orchestrator/pkg/database_support"
	"github.com/hexa-org/policy-orchestrator/pkg/hawk_support"
	"github.com/hexa-org/policy-orchestrator/pkg/orchestrator"
	"github.com/hexa-org/policy-orchestrator/pkg/web_support"
	"github.com/hexa-org/policy-orchestrator/pkg/workflow_support"
	"log"
	"net"
	"net/http"
	"os"
)

func App(key string, addr string, hostPort string, dbUrl string) (*http.Server, *workflow_support.WorkScheduler) {
	db, _ := database_support.Open(dbUrl)
	store := hawk_support.NewCredentialStore(key)
	handlers, scheduler := orchestrator.LoadHandlers(store, hostPort, db)
	return web_support.Create(addr, handlers, web_support.Options{}), scheduler
}

func newApp() (*http.Server, net.Listener, *workflow_support.WorkScheduler) {
	addr := "0.0.0.0:8885"
	if found := os.Getenv("PORT"); found != "" {
		addr = fmt.Sprintf("0.0.0.0:%v", found)
	}
	log.Printf("Found server address %v", addr)

	dbUrl := os.Getenv("POSTGRESQL_URL")
	key := os.Getenv("ORCHESTRATOR_KEY")
	hostPort := os.Getenv("ORCHESTRATOR_HOSTPORT")
	listener, _ := net.Listen("tcp", addr)
	app, scheduler := App(key, listener.Addr().String(), hostPort, dbUrl)
	return app, listener, scheduler
}

func main() {
	app, listener, scheduler := newApp()
	scheduler.Start()
	web_support.Start(app, listener)
}
