package admin

import (
	"github.com/gorilla/mux"
	"net/http"
)

type Client interface {
	Health(url string) (string, error)
	Applications(url string) ([]Application, error)
	Application(url string) (Application, error)
	Integrations(url string) ([]Integration, error)
	CreateIntegration(url string, provider string, key []byte) error
	DeleteIntegration(url string) error
	GetPolicies(url string) ([]Policy, string, error)
	SetPolicies(url string, policies string) error
}

func IndexHandler(w http.ResponseWriter, r *http.Request) {
	http.Redirect(w, r, "/integrations", http.StatusPermanentRedirect)
}

func LoadHandlers(orchestratorUrl string, client Client) func(router *mux.Router) {
	apps := NewApplicationsHandler(orchestratorUrl, client)
	integrations := NewIntegrationsHandler(orchestratorUrl, client)
	status := NewStatusHandler(orchestratorUrl, client)

	return func(router *mux.Router) {
		router.HandleFunc("/", IndexHandler).Methods("GET")
		router.HandleFunc("/applications", apps.List).Methods("GET")
		router.HandleFunc("/applications/{id}", apps.Show).Methods("GET")
		router.HandleFunc("/applications/{id}/edit", apps.Edit).Methods("GET")
		router.HandleFunc("/applications/{id}", apps.Update).Methods("POST")
		router.HandleFunc("/integrations", integrations.List).Methods("GET")
		router.HandleFunc("/integrations/new", integrations.New).Methods("GET")
		router.HandleFunc("/integrations", integrations.Create).Methods("POST")
		router.HandleFunc("/integrations/{id}", integrations.Delete).Methods("POST")
		router.HandleFunc("/status", status.StatusHandler).Methods("GET")

		fileServer := http.FileServer(http.Dir("pkg/admin/resources/static"))
		router.PathPrefix("/").Handler(http.StripPrefix("/", fileServer))
	}
}
