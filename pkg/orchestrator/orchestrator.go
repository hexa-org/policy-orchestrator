package orchestrator

import (
	"database/sql"
	"github.com/gorilla/mux"
	"github.com/hiyosi/hawk"
	"hexa/pkg/hawk_support"
	"hexa/pkg/workflow_support"
)

func LoadHandlers(store hawk.CredentialStore, hostPort string, database *sql.DB) (func(router *mux.Router), *workflow_support.WorkScheduler) {
	applicationsHandler := ApplicationsHandler{}
	gateway := IntegrationsDataGateway{database}
	integrationsHandler := IntegrationsHandler{gateway}

	worker := DiscoveryWorker{}
	finder := DiscoveryWorkFinder{Gateway: gateway}
	list := []workflow_support.Worker{&worker}
	scheduler := &workflow_support.WorkScheduler{Finder: &finder, Workers: list, Delay: 60_000}

	return func(router *mux.Router) {
		router.HandleFunc("/applications", hawk_support.HawkMiddleware(applicationsHandler.List, store, hostPort)).Methods("GET")
		router.HandleFunc("/integrations", hawk_support.HawkMiddleware(integrationsHandler.List, store, hostPort)).Methods("GET")
		router.HandleFunc("/integrations", hawk_support.HawkMiddleware(integrationsHandler.Create, store, hostPort)).Methods("POST")
		router.HandleFunc("/integrations/{id}", hawk_support.HawkMiddleware(integrationsHandler.Delete, store, hostPort)).Methods("GET")
	}, scheduler
}
