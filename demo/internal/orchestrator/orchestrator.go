package orchestrator

import (
	"github.com/gorilla/mux"
	"github.com/hexa-org/policy-mapper/api/policyprovider"
	"github.com/hexa-org/policy-orchestrator/demo/pkg/dataConfigGateway"
	"github.com/hexa-org/policy-orchestrator/demo/pkg/hawksupport"
	"github.com/hiyosi/hawk"
)

func LoadHandlers(configHandler *dataConfigGateway.ConfigData, store hawk.CredentialStore, hostPort string, cacheProviders map[string]policyprovider.Provider) func(router *mux.Router) {
	pb := NewProviderBuilder()
	if cacheProviders != nil {
		pb.AddProviders(cacheProviders)
	}
	integrationsGateway := configHandler
	applicationsGateway := configHandler.GetApplicationDataGateway()

	applicationsService := ApplicationsService{ApplicationsGateway: applicationsGateway, IntegrationsGateway: integrationsGateway, ProviderBuilder: pb}

	applicationsHandler := ApplicationsHandler{applicationsGateway, integrationsGateway, applicationsService}
	integrationsHandler := IntegrationsHandler{integrationsGateway}
	orchestrationHandler := OrchestrationHandler{applicationsService: applicationsService}

	return func(router *mux.Router) {
		router.HandleFunc("/applications", hawksupport.HawkMiddleware(applicationsHandler.List, store, hostPort)).Methods("GET")
		router.HandleFunc("/applications/{id}", hawksupport.HawkMiddleware(applicationsHandler.Show, store, hostPort)).Methods("GET")
		router.HandleFunc("/applications/{id}/policies", hawksupport.HawkMiddleware(applicationsHandler.GetPolicies, store, hostPort)).Methods("GET")
		router.HandleFunc("/applications/{id}/policies", hawksupport.HawkMiddleware(applicationsHandler.SetPolicies, store, hostPort)).Methods("POST")
		router.HandleFunc("/integrations", hawksupport.HawkMiddleware(integrationsHandler.List, store, hostPort)).Methods("GET")
		router.HandleFunc("/integrations", hawksupport.HawkMiddleware(integrationsHandler.Create, store, hostPort)).Methods("POST")
		router.HandleFunc("/integrations/{id}", hawksupport.HawkMiddleware(integrationsHandler.Delete, store, hostPort)).Methods("GET")
		router.HandleFunc("/orchestration", hawksupport.HawkMiddleware(orchestrationHandler.Update, store, hostPort)).Methods("POST")
	}
}
