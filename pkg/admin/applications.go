package admin

import (
	"bytes"
	"encoding/json"
	"fmt"
	"github.com/gorilla/mux"
	"github.com/hexa-org/policy-orchestrator/pkg/websupport"
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

type Policy struct {
	Version string
	Action  string
	Subject Subject
	Object  Object
}

type Subject struct {
	AuthenticatedUsers []string
}

type Object struct {
	Resources []string
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
		model := websupport.Model{Map: map[string]interface{}{"resource": "applications", "message": "Unable to contact orchestrator."}}
		_ = websupport.ModelAndView(w, "applications", model)
		log.Println(err)
		return
	}
	model := websupport.Model{Map: map[string]interface{}{"resource": "applications", "applications": applications}}
	_ = websupport.ModelAndView(w, "applications", model)
}

func (p applicationsHandler) Show(w http.ResponseWriter, r *http.Request) {
	identifier := mux.Vars(r)["id"]
	app, err := p.client.Application(fmt.Sprintf("%v/applications/%s", p.orchestratorUrl, identifier))
	policies, rawJson, anotherPossibleErr := p.client.Policies(fmt.Sprintf("%v/applications/%s/policies", p.orchestratorUrl, identifier))
	if err != nil || anotherPossibleErr != nil {
		model := websupport.Model{Map: map[string]interface{}{"resource": "applications", "message": "Unable to contact orchestrator."}}
		_ = websupport.ModelAndView(w, "applications_show", model)
		log.Println(err)
		return
	}
	var buffer bytes.Buffer
	_ = json.Indent(&buffer, []byte(rawJson), "", "  ")
	model := websupport.Model{Map: map[string]interface{}{"resource": "applications", "application": app, "policies": policies, "rawJson": buffer.String()}}
	_ = websupport.ModelAndView(w, "applications_show", model)
}
