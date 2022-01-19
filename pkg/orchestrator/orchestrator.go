package orchestrator

import (
	"database/sql"
	"github.com/gorilla/mux"
	"github.com/hexa-org/policy-orchestrator/pkg/hawksupport"
	"github.com/hexa-org/policy-orchestrator/pkg/orchestrator/provider"
	"github.com/hexa-org/policy-orchestrator/pkg/workflowsupport"
	"github.com/hiyosi/hawk"
)

func LoadHandlers(providers map[string]provider.Provider, store hawk.CredentialStore, hostPort string, database *sql.DB) (func(router *mux.Router), *workflowsupport.WorkScheduler) {
	applicationsGateway := ApplicationsDataGateway{database}
	integrationsGateway := IntegrationsDataGateway{database}

	worker := DiscoveryWorker{providers, applicationsGateway}
	finder := NewDiscoveryWorkFinder(integrationsGateway)

	applicationsHandler := ApplicationsHandler{applicationsGateway, integrationsGateway, providers}
	integrationsHandler := IntegrationsHandler{integrationsGateway, worker}

	list := []workflowsupport.Worker{&worker}
	scheduler := workflowsupport.NewScheduler(&finder, list, 60_000)

	Report(finder.Results)

	return func(router *mux.Router) {
		router.HandleFunc("/applications", hawksupport.HawkMiddleware(applicationsHandler.List, store, hostPort)).Methods("GET")
		router.HandleFunc("/applications/{id}", hawksupport.HawkMiddleware(applicationsHandler.Show, store, hostPort)).Methods("GET")
		router.HandleFunc("/applications/{id}/policies", hawksupport.HawkMiddleware(applicationsHandler.GetPolicies, store, hostPort)).Methods("GET")
		router.HandleFunc("/integrations", hawksupport.HawkMiddleware(integrationsHandler.List, store, hostPort)).Methods("GET")
		router.HandleFunc("/integrations", hawksupport.HawkMiddleware(integrationsHandler.Create, store, hostPort)).Methods("POST")
		router.HandleFunc("/integrations/{id}", hawksupport.HawkMiddleware(integrationsHandler.Delete, store, hostPort)).Methods("GET")
	}, &scheduler
}
