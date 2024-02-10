package orchestrator

import (
	"database/sql"

	"github.com/gorilla/mux"
	"github.com/hexa-org/policy-mapper/api/policyprovider"
	"github.com/hexa-org/policy-orchestrator/demo/pkg/hawksupport"
	"github.com/hexa-org/policy-orchestrator/demo/pkg/workflowsupport"
	"github.com/hiyosi/hawk"
)

func LoadHandlers(database *sql.DB, store hawk.CredentialStore, hostPort string, legacyProviders map[string]policyprovider.Provider) (func(router *mux.Router), *workflowsupport.WorkScheduler) {
	pb := NewProviderBuilder(legacyProviders)
	applicationsGateway := ApplicationsDataGateway{database}
	integrationsGateway := IntegrationsDataGateway{database}
	applicationsService := ApplicationsService{ApplicationsGateway: applicationsGateway, IntegrationsGateway: integrationsGateway, ProviderBuilder: pb}

	worker := DiscoveryWorker{pb, applicationsGateway}
	finder := NewDiscoveryWorkFinder(integrationsGateway)

	applicationsHandler := ApplicationsHandler{applicationsGateway, integrationsGateway, applicationsService}
	integrationsHandler := IntegrationsHandler{integrationsGateway, worker}
	orchestrationHandler := OrchestrationHandler{applicationsService: applicationsService}

	list := []workflowsupport.Worker{&worker}
	scheduler := workflowsupport.NewScheduler(&finder, list, 60_000)

	return func(router *mux.Router) {
		router.HandleFunc("/applications", hawksupport.HawkMiddleware(applicationsHandler.List, store, hostPort)).Methods("GET")
		router.HandleFunc("/applications/{id}", hawksupport.HawkMiddleware(applicationsHandler.Show, store, hostPort)).Methods("GET")
		router.HandleFunc("/applications/{id}/policies", hawksupport.HawkMiddleware(applicationsHandler.GetPolicies, store, hostPort)).Methods("GET")
		router.HandleFunc("/applications/{id}/policies", hawksupport.HawkMiddleware(applicationsHandler.SetPolicies, store, hostPort)).Methods("POST")
		router.HandleFunc("/integrations", hawksupport.HawkMiddleware(integrationsHandler.List, store, hostPort)).Methods("GET")
		router.HandleFunc("/integrations", hawksupport.HawkMiddleware(integrationsHandler.Create, store, hostPort)).Methods("POST")
		router.HandleFunc("/integrations/{id}", hawksupport.HawkMiddleware(integrationsHandler.Delete, store, hostPort)).Methods("GET")
		router.HandleFunc("/orchestration", hawksupport.HawkMiddleware(orchestrationHandler.Update, store, hostPort)).Methods("POST")
	}, &scheduler
}
