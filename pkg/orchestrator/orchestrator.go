package orchestrator

import (
	"database/sql"
	"github.com/gorilla/mux"
	"github.com/hiyosi/hawk"
	"hexa/pkg/hawk_support"
)

func LoadHandlers(store hawk.CredentialStore, hostPort string, database *sql.DB) func(router *mux.Router) {
	applicationsHandler := ApplicationsHandler{}
	gateway := IntegrationsDataGateway{database}
	integrationsHandler := IntegrationsHandler{gateway}
	return func(router *mux.Router) {
		router.HandleFunc("/applications", hawk_support.HawkMiddleware(applicationsHandler.List, store, hostPort)).Methods("GET")
		router.HandleFunc("/integrations", hawk_support.HawkMiddleware(integrationsHandler.List, store, hostPort)).Methods("GET")
		router.HandleFunc("/integrations", hawk_support.HawkMiddleware(integrationsHandler.Create, store, hostPort)).Methods("POST")
		router.HandleFunc("/integrations/{id}", hawk_support.HawkMiddleware(integrationsHandler.Delete, store, hostPort)).Methods("GET")
	}
}