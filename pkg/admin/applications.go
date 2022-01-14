package admin

import (
	"fmt"
	"github.com/gorilla/mux"
	"github.com/hexa-org/policy-orchestrator/pkg/web_support"
	"log"
	"net/http"
)

type Application struct {
	ID            string
	IntegrationId string
	ObjectId      string
	Name          string
	Description   string
}

type applicationsHandler struct {
	orchestratorUrl string
	client          Client
}

func NewApplicationsHandler(orchestratorUrl string, client Client) applicationsHandler {
	return applicationsHandler{orchestratorUrl, client}
}

func (p applicationsHandler) List(w http.ResponseWriter, r *http.Request) {
	url := fmt.Sprintf("%v/applications", p.orchestratorUrl)
	applications, err := p.client.Applications(url)
	if err != nil {
		model := web_support.Model{Map: map[string]interface{}{"resource": "applications", "message": "Unable to contact orchestrator."}}
		_ = web_support.ModelAndView(w, "applications", model)
		log.Println(err)
		return
	}
	model := web_support.Model{Map: map[string]interface{}{"resource": "applications", "applications": applications}}
	_ = web_support.ModelAndView(w, "applications", model)
}

func (p applicationsHandler) Show(w http.ResponseWriter, r *http.Request) {
	identifier := mux.Vars(r)["id"]
	url := fmt.Sprintf("%v/applications/%s", p.orchestratorUrl, identifier)
	app, err := p.client.Application(url)
	if err != nil {
		model := web_support.Model{Map: map[string]interface{}{"resource": "applications", "message": "Unable to contact orchestrator."}}
		_ = web_support.ModelAndView(w, "applications_show", model)
		log.Println(err)
		return
	}
	model := web_support.Model{Map: map[string]interface{}{"resource": "applications", "application": app}}
	_ = web_support.ModelAndView(w, "applications_show", model)
}
