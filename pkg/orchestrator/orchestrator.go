package orchestrator

import (
	"github.com/gorilla/mux"
	"github.com/hiyosi/hawk"
	"hexa/pkg/hawk_support"
)

func LoadHandlers(store hawk.CredentialStore, hostPort string) func(router *mux.Router) {
	handler := NewApplicationsHandler()
	return func(router *mux.Router) {
		router.HandleFunc("/applications", hawk_support.HawkMiddleware(handler.Applications, store, hostPort)).Methods("GET")
	}
}