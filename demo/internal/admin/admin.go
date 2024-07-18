package admin

import (
	"embed"
	"io/fs"
	"net/http"

	"github.com/gorilla/mux"
	"github.com/hexa-org/policy-mapper/pkg/hexapolicy"
	"github.com/hexa-org/policy-mapper/pkg/oidcSupport"
	"github.com/hexa-org/policy-mapper/pkg/sessionSupport"
	log "golang.org/x/exp/slog"
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
	GetHttpClient() HTTPClient
}

func IndexHandler(w http.ResponseWriter, r *http.Request) {
	http.Redirect(w, r, "/integrations", http.StatusPermanentRedirect)
}

func LoadHandlers(orchestratorUrl string, client Client, sessionHandler sessionSupport.SessionManager) func(router *mux.Router) {

	oidcHandler, err := oidcSupport.NewOidcClientHandler(sessionHandler, &resources)
	apps := NewApplicationsHandler(sessionHandler, orchestratorUrl, client)

	integrations := NewIntegrationsHandler(orchestratorUrl, client, sessionHandler)
	orchestration := NewOrchestrationHandler(orchestratorUrl, client, sessionHandler)
	status := NewStatusHandler(orchestratorUrl, client)

	return func(router *mux.Router) {

		oidcHandler.InitHandlers(router)
		if err != nil {
			log.Error(err.Error())
			log.Warn("OIDC Login is disabled")
		}
		if !oidcHandler.Enabled {
			router.HandleFunc("/", IndexHandler)
		}

		router.HandleFunc("/integrations", oidcHandler.HandleSessionScope(integrations.List, []string{"integration"})).Methods("GET")
		router.HandleFunc("/integrations/new", oidcHandler.HandleSessionScope(integrations.New, []string{"integration"})).Methods("GET").Queries("provider", "{provider}")
		router.HandleFunc("/integrations", oidcHandler.HandleSessionScope(integrations.CreateIntegration, []string{"integration"})).Methods("POST")
		router.HandleFunc("/integrations/{id}", oidcHandler.HandleSessionScope(integrations.Delete, []string{"integration"})).Methods("POST")
		router.HandleFunc("/applications", oidcHandler.HandleSessionScope(apps.List, []string{"integration"})).Methods("GET")
		router.HandleFunc("/applications/{id}", oidcHandler.HandleSessionScope(apps.Show, []string{"integration"})).Methods("GET")
		router.HandleFunc("/applications/{id}/policies", oidcHandler.HandleSessionScope(apps.Policies, []string{"integration"})).Methods("GET")
		router.HandleFunc("/applications/{id}/edit", oidcHandler.HandleSessionScope(apps.Edit, []string{"integration"})).Methods("GET")
		router.HandleFunc("/applications/{id}", oidcHandler.HandleSessionScope(apps.Update, []string{"integration"})).Methods("POST")
		router.HandleFunc("/orchestration/new", oidcHandler.HandleSessionScope(orchestration.New, []string{"integration"})).Methods("GET")
		router.HandleFunc("/orchestration", oidcHandler.HandleSessionScope(orchestration.Update, []string{"integration"})).Methods("POST")
		router.HandleFunc("/status", oidcHandler.HandleSessionScope(status.StatusHandler, []string{"integration"})).Methods("GET")

		staticFs, _ := fs.Sub(resources, "resources/static")
		fileServer := http.FileServer(http.FS(staticFs))
		router.PathPrefix("/").Handler(http.StripPrefix("/", fileServer))
	}

}
