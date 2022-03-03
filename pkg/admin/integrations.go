package admin

import (
	"bytes"
	"encoding/json"
	"fmt"
	"github.com/gorilla/mux"
	"github.com/hexa-org/policy-orchestrator/pkg/websupport"
	"io/ioutil"
	"log"
	"net/http"
	"strings"
)

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
}

func NewIntegrationsHandler(orchestratorUrl string, client Client) IntegrationHandler {
	return integrationsHandler{orchestratorUrl, client}
}

func (i integrationsHandler) List(w http.ResponseWriter, r *http.Request) {
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

type googleKeyFile struct {
	ProjectId string `json:"project_id"`
}

type azureKeyFile struct {
	Tenant string `json:"tenant"`
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

	// todo - replace conditional logic with strategy or new route
	if provider == "google_cloud" {
		var foundKeyFile googleKeyFile
		err = json.NewDecoder(bytes.NewReader(key)).Decode(&foundKeyFile)
		if err != nil || foundKeyFile.ProjectId == "" {
			log.Println("Unable to read key file.")
			model := websupport.Model{Map: map[string]interface{}{"resource": "integrations", "provider": provider, "message": "Unable to read key file."}}
			_ = websupport.ModelAndView(w, integrationView, model)
			return
		}
		name = foundKeyFile.ProjectId
	}

	if provider == "azure" {
		var foundKeyFile azureKeyFile
		err = json.NewDecoder(bytes.NewReader(key)).Decode(&foundKeyFile)
		if err != nil || foundKeyFile.Tenant == "" {
			log.Println("Unable to read key file.")
			model := websupport.Model{Map: map[string]interface{}{"resource": "integrations", "provider": provider, "message": "Unable to read key file."}}
			_ = websupport.ModelAndView(w, integrationView, model)
			return
		}
		name = foundKeyFile.Tenant
	}

	if provider == "amazon" {
		name = "amazon"
	}

	err = i.client.CreateIntegration(url, name, provider, key)
	if err != nil {
		log.Println("Unable to communicate with orchestrator.")
		model := websupport.Model{Map: map[string]interface{}{"resource": "integrations", "provider": provider, "message": "Unable to communicate with orchestrator."}}
		_ = websupport.ModelAndView(w, integrationView, model)
		return
	}
	http.Redirect(w, r, "/integrations", http.StatusMovedPermanently)
}

type azureKey struct {
	AppId        string `json:"appId"`
	Secret       string `json:"secret"`
	Tenant       string `json:"tenant"`
	Subscription string `json:"subscription"`
}

func (i integrationsHandler) knownIntegrationViews(provider string) string {
	integrationViews := make(map[string]string)
	integrationViews["google_cloud"] = "integrations_new_google_cloud"
	integrationViews["azure"] = "integrations_new_azure"
	integrationViews["amazon"] = "integrations_new_amazon"
	integrationView := integrationViews[strings.ToLower(provider)]
	return integrationView
}

func (i integrationsHandler) Delete(w http.ResponseWriter, r *http.Request) {
	url := fmt.Sprintf("%v/integrations/%s", i.orchestratorUrl, mux.Vars(r)["id"])
	err := i.client.DeleteIntegration(url)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
	http.Redirect(w, r, "/integrations", http.StatusMovedPermanently)
}
