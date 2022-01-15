package main

import (
	"fmt"
	"github.com/hexa-org/policy-orchestrator/pkg/databasesupport"
	"github.com/hexa-org/policy-orchestrator/pkg/hawksupport"
	"github.com/hexa-org/policy-orchestrator/pkg/orchestrator"
	"github.com/hexa-org/policy-orchestrator/pkg/websupport"
	"github.com/hexa-org/policy-orchestrator/pkg/workflowsupport"
	"log"
	"net"
	"net/http"
	"os"
)

func App(key string, addr string, hostPort string, dbUrl string) (*http.Server, *workflowsupport.WorkScheduler) {
	db, _ := databasesupport.Open(dbUrl)
	store := hawksupport.NewCredentialStore(key)
	handlers, scheduler := orchestrator.LoadHandlers(store, hostPort, db)
	return websupport.Create(addr, handlers, websupport.Options{}), scheduler
}

func newApp(addr string) (*http.Server, net.Listener, *workflowsupport.WorkScheduler) {
	if found := os.Getenv("PORT"); found != "" {
		host, _, _ := net.SplitHostPort(addr)
		addr = fmt.Sprintf("%v:%v", host, found)
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
	app, listener, scheduler := newApp("0.0.0.0:8885")
	scheduler.Start()
	websupport.Start(app, listener)
}
