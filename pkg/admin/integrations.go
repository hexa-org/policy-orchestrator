package admin

import (
	"fmt"
	"github.com/gorilla/mux"
	"hexa/pkg/web_support"
	"io/ioutil"
	"net/http"
)

type Integration struct {
	ID       string
	Name     string
	Provider string
	Key      []byte
}

type integrationsHandler struct {
	orchestratorUrl string
	client          Client
}

func NewIntegrationsHandler(orchestratorUrl string, client Client) integrationsHandler {
	return integrationsHandler{orchestratorUrl, client}
}

func (i integrationsHandler) List(w http.ResponseWriter, r *http.Request) {
	url := fmt.Sprintf("%v/integrations", i.orchestratorUrl)
	integrations, _ := i.client.Integrations(url)
	model := web_support.Model{Map: map[string]interface{}{"resource": "integrations", "integrations": integrations}}
	_ = web_support.ModelAndView(w, "integrations", model)
}

func (i integrationsHandler) New(w http.ResponseWriter, r *http.Request) {
	model := web_support.Model{Map: map[string]interface{}{"resource": "integrations"}}
	_ = web_support.ModelAndView(w, "integration_new", model)
}

func (i integrationsHandler) Create(w http.ResponseWriter, r *http.Request) {
	url := fmt.Sprintf("%v/integrations", i.orchestratorUrl)

	err := r.ParseMultipartForm(32 << 20)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	provider := r.FormValue("provider")
	file, _, err := r.FormFile("key")
	if err != nil {
		return
	}

	var key []byte
	key, err = ioutil.ReadAll(file)
	if err != nil {
		return
	}
	_ = file.Close()

	err = i.client.CreateIntegration(url, provider, key)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
	http.Redirect(w, r, "/integrations", 301)
}

func (i integrationsHandler) Delete(w http.ResponseWriter, r *http.Request) {
	url := fmt.Sprintf("%v/integrations/%s", i.orchestratorUrl, mux.Vars(r)["id"])
	err := i.client.DeleteIntegration(url)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
	http.Redirect(w, r, "/integrations", 301)
}
