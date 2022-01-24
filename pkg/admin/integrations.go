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
	CreateGoogleIntegration(w http.ResponseWriter, r *http.Request)
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
	model := websupport.Model{Map: map[string]interface{}{"resource": "integrations"}}
	_ = websupport.ModelAndView(w, "integrations_new", model)
}

type keyFile struct {
	ProjectId string `json:"project_id"`
}

func (i integrationsHandler) CreateGoogleIntegration(w http.ResponseWriter, r *http.Request) {
	url := fmt.Sprintf("%v/integrations", i.orchestratorUrl)

	err := r.ParseMultipartForm(32 << 20)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	provider := r.FormValue("provider")
	file, _, err := r.FormFile("key")
	if err != nil {
		log.Printf("Missing key file %s.\n", err.Error())
		model := websupport.Model{Map: map[string]interface{}{"resource": "integrations", "message": "Missing key file."}}
		_ = websupport.ModelAndView(w, "integrations_new", model)
		return
	}

	var key []byte
	key, err = ioutil.ReadAll(file)
	if err != nil {
		return
	}
	_ = file.Close()

	var foundKeyFile keyFile
	err = json.NewDecoder(bytes.NewReader(key)).Decode(&foundKeyFile)
	if err != nil || foundKeyFile.ProjectId == "" {
		log.Println("Unable to read key file.")
		model := websupport.Model{Map: map[string]interface{}{"resource": "integrations", "message": "Unable to read key file."}}
		_ = websupport.ModelAndView(w, "integrations_new", model)
		return
	}
	name := foundKeyFile.ProjectId

	err = i.client.CreateIntegration(url, name, provider, key)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
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
