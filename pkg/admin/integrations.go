package admin

import (
	"fmt"
	"github.com/gorilla/mux"
	"github.com/hexa-org/policy-orchestrator/pkg/websupport"
	"io/ioutil"
	"log"
	"net/http"
	"strings"
)

type IntegrationProviderInterface interface {
	detect(provider string) bool
	name(key []byte) (string, error)
}

type Integration struct {
	ID       string
	Name     string
	Provider string
	Key      []byte
}

type IntegrationHandler interface {
	List(w http.ResponseWriter, r *http.Request)
	New(w http.ResponseWriter, r *http.Request)
	CreateIntegration(w http.ResponseWriter, r *http.Request)
	Delete(w http.ResponseWriter, r *http.Request)
}

type integrationsHandler struct {
	orchestratorUrl string
	client          Client
	providerStructs []IntegrationProviderInterface
}

func NewIntegrationsHandler(orchestratorUrl string, client Client) IntegrationHandler {
	return integrationsHandler{
		orchestratorUrl,
		client,
		[]IntegrationProviderInterface{googleProvider{}, azureProvider{}, amazonProvider{}, opaProvider{}},
	}
}

func (i integrationsHandler) List(w http.ResponseWriter, _ *http.Request) {
	url := fmt.Sprintf("%v/integrations", i.orchestratorUrl)
	integrations, err := i.client.Integrations(url)
	if err != nil {
		model := websupport.Model{Map: map[string]interface{}{"resource": "integrations", "message": "Unable to contact orchestrator."}}
		_ = websupport.ModelAndView(w, "integrations", model)
		log.Println(err)
		return
	}
	model := websupport.Model{Map: map[string]interface{}{"resource": "integrations", "integrations": integrations}}
	_ = websupport.ModelAndView(w, "integrations", model)
}

func (i integrationsHandler) New(w http.ResponseWriter, r *http.Request) {
	provider := r.URL.Query().Get("provider")
	model := websupport.Model{Map: map[string]interface{}{"resource": "integrations", "provider": provider}}
	integrationView := i.knownIntegrationViews(provider)
	_ = websupport.ModelAndView(w, integrationView, model)
}

func (i integrationsHandler) CreateIntegration(w http.ResponseWriter, r *http.Request) {
	url := fmt.Sprintf("%v/integrations", i.orchestratorUrl)

	err := r.ParseMultipartForm(32 << 20)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	provider := r.FormValue("provider")
	integrationView := i.knownIntegrationViews(provider)

	var key []byte
	var name string

	file, _, err := r.FormFile("key")
	if err != nil {
		log.Printf("Missing key file %s.\n", err.Error())
		model := websupport.Model{Map: map[string]interface{}{"resource": "integrations", "provider": provider, "message": "Missing key file."}}
		_ = websupport.ModelAndView(w, integrationView, model)
		return
	}

	key, err = ioutil.ReadAll(file)
	if err != nil {
		return
	}
	_ = file.Close()

	var foundProvider IntegrationProviderInterface
	for _, p := range i.providerStructs {
		if p.detect(provider) {
			foundProvider = p
		}
	}
	if foundProvider == nil {
		i.viewWithMessage(w, provider, "unknown provider", integrationView)
		return
	}
	name, err = foundProvider.name(key)
	if err != nil {
		i.viewWithMessage(w, provider, err.Error(), integrationView)
		return
	}

	err = i.client.CreateIntegration(url, name, provider, key)
	if err != nil {
		model := websupport.Model{Map: map[string]interface{}{"resource": "integrations", "provider": provider, "message": "Unable to communicate with orchestrator."}}
		_ = websupport.ModelAndView(w, integrationView, model)
		return
	}
	http.Redirect(w, r, "/integrations", http.StatusMovedPermanently)
}

func (i integrationsHandler) Delete(w http.ResponseWriter, r *http.Request) {
	url := fmt.Sprintf("%v/integrations/%s", i.orchestratorUrl, mux.Vars(r)["id"])
	err := i.client.DeleteIntegration(url)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
	http.Redirect(w, r, "/integrations", http.StatusMovedPermanently)
}

func (i integrationsHandler) viewWithMessage(w http.ResponseWriter, provider string, message string, integrationView string) {
	model := websupport.Model{Map: map[string]interface{}{"resource": "integrations", "provider": provider, "message": message}}
	_ = websupport.ModelAndView(w, integrationView, model)
}

func (i integrationsHandler) knownIntegrationViews(provider string) string {
	integrationViews := make(map[string]string)
	integrationViews["google_cloud"] = "integrations_new_google_cloud"
	integrationViews["azure"] = "integrations_new_azure"
	integrationViews["amazon"] = "integrations_new_amazon"
	integrationViews["open_policy_agent"] = "integrations_new_open_policy"
	integrationView := integrationViews[strings.ToLower(provider)]
	return integrationView
}
