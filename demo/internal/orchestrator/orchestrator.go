package orchestrator

import (
	"github.com/gorilla/mux"
	"github.com/hexa-org/policy-mapper/api/policyprovider"
	"github.com/hexa-org/policy-orchestrator/demo/pkg/dataConfigGateway"
	"github.com/hexa-org/policy-orchestrator/demo/pkg/oauth2support"
)

func LoadHandlers(configHandler *dataConfigGateway.ConfigData, cacheProviders map[string]policyprovider.Provider) func(router *mux.Router) {
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
	jwtHandler := oauth2support.NewResourceServerJwtHandler()

	return func(router *mux.Router) {
		router.HandleFunc("/applications", oauth2support.JwtAuthenticationHandler(applicationsHandler.List, jwtHandler)).Methods("GET")
		router.HandleFunc("/applications/{id}", oauth2support.JwtAuthenticationHandler(applicationsHandler.Show, jwtHandler)).Methods("GET")
		router.HandleFunc("/applications/{id}/policies", oauth2support.JwtAuthenticationHandler(applicationsHandler.GetPolicies, jwtHandler)).Methods("GET")
		router.HandleFunc("/applications/{id}/policies", oauth2support.JwtAuthenticationHandler(applicationsHandler.SetPolicies, jwtHandler)).Methods("POST")
		router.HandleFunc("/integrations", oauth2support.JwtAuthenticationHandler(integrationsHandler.List, jwtHandler)).Methods("GET")
		router.HandleFunc("/integrations", oauth2support.JwtAuthenticationHandler(integrationsHandler.Create, jwtHandler)).Methods("POST")
		router.HandleFunc("/integrations/{id}", oauth2support.JwtAuthenticationHandler(integrationsHandler.Delete, jwtHandler)).Methods("GET")
		router.HandleFunc("/orchestration", oauth2support.JwtAuthenticationHandler(orchestrationHandler.Update, jwtHandler)).Methods("POST")
	}
}
