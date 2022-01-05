package admin

import (
	"github.com/gorilla/mux"
	"net/http"
)

type Client interface {
	Health(url string) (string, error)
	Applications(url string) ([]Application, error)
}

///

func IndexHandler(w http.ResponseWriter, r *http.Request) {
	http.Redirect(w, r, "/discovery", 301)
}

func LoadHandlers(orchestratorUrl string, client Client) func(router *mux.Router) {
	apps := NewApplicationsHandler(orchestratorUrl, client)
	integrations := NewIntegrationsHandler()
	status := NewStatusHandler(orchestratorUrl, client)

	return func(router *mux.Router) {
		router.HandleFunc("/", IndexHandler).Methods("GET")
		router.HandleFunc("/applications", apps.ApplicationsHandler).Methods("GET")
		router.HandleFunc("/discovery", integrations.IntegrationsHandler).Methods("GET")
		router.HandleFunc("/status", status.StatusHandler).Methods("GET")

		fileServer := http.FileServer(http.Dir("pkg/admin/resources/static"))
		router.PathPrefix("/").Handler(http.StripPrefix("/", fileServer))
	}
}
