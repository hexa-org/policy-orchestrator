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

func newApp() (*http.Server, net.Listener, *workflowsupport.WorkScheduler) {
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
	websupport.Start(app, listener)
}
