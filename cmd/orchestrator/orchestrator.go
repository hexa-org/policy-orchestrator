package main

import (
	"fmt"
	"hexa/pkg/database_support"
	"hexa/pkg/hawk_support"
	"hexa/pkg/orchestrator"
	"hexa/pkg/web_support"
	"hexa/pkg/workflow_support"
	"log"
	"net/http"
	"os"
)

func App(key string, addr string, hostPort string, dbUrl string) (*http.Server, *workflow_support.WorkScheduler) {
	db, _ := database_support.Open(dbUrl)
	store := hawk_support.NewCredentialStore(key)
	handlers, scheduler := orchestrator.LoadHandlers(store, hostPort, db)
	return web_support.Create(addr, handlers, web_support.Options{}), scheduler
}

func newApp() (*http.Server, *workflow_support.WorkScheduler) {
	addr := "0.0.0.0:8885"
	if found := os.Getenv("PORT"); found != "" {
		addr = fmt.Sprintf("0.0.0.0:%v", found)
	}
	log.Printf("Found server address %v", addr)

	dbUrl := os.Getenv("POSTGRESQL_URL")
	key := os.Getenv("ORCHESTRATOR_KEY")
	hostPort := os.Getenv("ORCHESTRATOR_HOSTPORT")
	app, scheduler := App(key, addr, hostPort, dbUrl)
	return app, scheduler
}

func main() {
	app, scheduler := newApp()
	scheduler.Start()
	web_support.Start(app)
}
