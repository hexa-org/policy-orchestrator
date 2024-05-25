package admin

import (
	"embed"
	"io/fs"
	"net/http"

	"github.com/gorilla/mux"
	"github.com/hexa-org/policy-mapper/pkg/hexapolicy"
)

//go:embed resources
var resources embed.FS

type Client interface {
	Health() (string, error)
	Integrations() ([]Integration, error)
	CreateIntegration(name string, provider string, key []byte) error
	DeleteIntegration(id string) error
	Applications(refresh bool) ([]Application, error)
	Application(id string) (Application, error)
	GetPolicies(id string) ([]hexapolicy.PolicyInfo, string, error)
	SetPolicies(id string, policies string) error
	Orchestration(from string, to string) error
}

func IndexHandler(w http.ResponseWriter, r *http.Request) {
	http.Redirect(w, r, "/integrations", http.StatusPermanentRedirect)
}

func LoadHandlers(orchestratorUrl string, client Client) func(router *mux.Router) {
	apps := NewApplicationsHandler(orchestratorUrl, client)
	integrations := NewIntegrationsHandler(orchestratorUrl, client)
	orchestration := NewOrchestrationHandler(orchestratorUrl, client)
	status := NewStatusHandler(orchestratorUrl, client)

	return func(router *mux.Router) {
		router.HandleFunc("/", IndexHandler).Methods("GET")
		router.HandleFunc("/integrations", integrations.List).Methods("GET")
		router.HandleFunc("/integrations/new", integrations.New).Methods("GET").Queries("provider", "{provider}")
		router.HandleFunc("/integrations", integrations.CreateIntegration).Methods("POST")
		router.HandleFunc("/integrations/{id}", integrations.Delete).Methods("POST")
		router.HandleFunc("/applications", apps.List).Methods("GET")
		router.HandleFunc("/applications/{id}", apps.Show).Methods("GET")
		router.HandleFunc("/applications/{id}/policies", apps.Policies).Methods("GET")
		router.HandleFunc("/applications/{id}/edit", apps.Edit).Methods("GET")
		router.HandleFunc("/applications/{id}", apps.Update).Methods("POST")
		router.HandleFunc("/orchestration/new", orchestration.New).Methods("GET")
		router.HandleFunc("/orchestration", orchestration.Update).Methods("POST")
		router.HandleFunc("/status", status.StatusHandler).Methods("GET")

		staticFs, _ := fs.Sub(resources, "resources/static")
		fileServer := http.FileServer(http.FS(staticFs))
		router.PathPrefix("/").Handler(http.StripPrefix("/", fileServer))
	}
}
