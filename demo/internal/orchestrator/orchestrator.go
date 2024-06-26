package orchestrator

import (
	"github.com/gorilla/mux"
	"github.com/hexa-org/policy-mapper/api/policyprovider"
	"github.com/hexa-org/policy-mapper/pkg/oauth2support"
	"github.com/hexa-org/policy-orchestrator/demo/pkg/dataConfigGateway"
	log "golang.org/x/exp/slog"
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
	jwtHandler, err := oauth2support.NewResourceJwtAuthorizer()
	if err != nil {
		log.Error("Error initializing JWT authorizer", "err", err.Error())
	}

	scopes := []string{"orchestrator"}

	return func(router *mux.Router) {
		router.HandleFunc("/applications", oauth2support.JwtAuthenticationHandler(applicationsHandler.List, jwtHandler, scopes)).Methods("GET")
		router.HandleFunc("/applications/{id}", oauth2support.JwtAuthenticationHandler(applicationsHandler.Show, jwtHandler, scopes)).Methods("GET")
		router.HandleFunc("/applications/{id}/policies", oauth2support.JwtAuthenticationHandler(applicationsHandler.GetPolicies, jwtHandler, scopes)).Methods("GET")
		router.HandleFunc("/applications/{id}/policies", oauth2support.JwtAuthenticationHandler(applicationsHandler.SetPolicies, jwtHandler, scopes)).Methods("POST")
		router.HandleFunc("/integrations", oauth2support.JwtAuthenticationHandler(integrationsHandler.List, jwtHandler, scopes)).Methods("GET")
		router.HandleFunc("/integrations", oauth2support.JwtAuthenticationHandler(integrationsHandler.Create, jwtHandler, scopes)).Methods("POST")
		router.HandleFunc("/integrations/{id}", oauth2support.JwtAuthenticationHandler(integrationsHandler.Delete, jwtHandler, scopes)).Methods("GET")
		router.HandleFunc("/orchestration", oauth2support.JwtAuthenticationHandler(orchestrationHandler.Update, jwtHandler, scopes)).Methods("POST")
	}
}
