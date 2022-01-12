package orchestrator

import (
	"database/sql"
	"github.com/gorilla/mux"
	"github.com/hiyosi/hawk"
	"hexa/pkg/hawk_support"
	"hexa/pkg/orchestrator/provider"
	"hexa/pkg/providers/google_cloud"
	"hexa/pkg/workflow_support"
)

func LoadHandlers(store hawk.CredentialStore, hostPort string, database *sql.DB) (func(router *mux.Router), *workflow_support.WorkScheduler) {
	applicationsGateway := ApplicationsDataGateway{database}
	integrationsGateway := IntegrationsDataGateway{database}

	googleProvider := &google_cloud.GoogleProvider{}
	worker := DiscoveryWorker{[]provider.Provider{googleProvider}, applicationsGateway}
	finder := DiscoveryWorkFinder{Gateway: integrationsGateway}

	applicationsHandler := ApplicationsHandler{applicationsGateway}
	integrationsHandler := IntegrationsHandler{integrationsGateway, worker}

	list := []workflow_support.Worker{&worker}
	scheduler := workflow_support.NewScheduler(&finder, list, 60_000)

	return func(router *mux.Router) {
		router.HandleFunc("/applications", hawk_support.HawkMiddleware(applicationsHandler.List, store, hostPort)).Methods("GET")
		router.HandleFunc("/applications/{id}", hawk_support.HawkMiddleware(applicationsHandler.Show, store, hostPort)).Methods("GET")
		router.HandleFunc("/integrations", hawk_support.HawkMiddleware(integrationsHandler.List, store, hostPort)).Methods("GET")
		router.HandleFunc("/integrations", hawk_support.HawkMiddleware(integrationsHandler.Create, store, hostPort)).Methods("POST")
		router.HandleFunc("/integrations/{id}", hawk_support.HawkMiddleware(integrationsHandler.Delete, store, hostPort)).Methods("GET")
	}, &scheduler
}
