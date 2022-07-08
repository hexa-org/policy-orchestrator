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
	ProviderName  string
}

type Policy struct {
	Meta    Meta     `validate:"required"`
	Actions []Action `validate:"required"`
	Subject Subject  `validate:"required"`
	Object  Object   `validate:"required"`
}

type Meta struct {
	Version string `validate:"required"`
}

type Action struct {
	ActionUri string `validate:"required"`
}

type Subject struct {
	Members []string `validate:"required"`
}

type Object struct {
	ResourceID string `validate:"required"`
}

type ApplicationsHandler interface {
	List(w http.ResponseWriter, r *http.Request)
	Show(w http.ResponseWriter, r *http.Request)
	Edit(w http.ResponseWriter, r *http.Request)
	Update(w http.ResponseWriter, r *http.Request)
}

type appsHandler struct {
	orchestratorUrl string
	client          Client
}

func NewApplicationsHandler(orchestratorUrl string, client Client) ApplicationsHandler {
	return appsHandler{orchestratorUrl, client}
}

func (p appsHandler) List(w http.ResponseWriter, _ *http.Request) {
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

func (p appsHandler) Show(w http.ResponseWriter, r *http.Request) {
	identifier := mux.Vars(r)["id"]
	orchestratorAppEndpoint := fmt.Sprintf("%v/applications/%s", p.orchestratorUrl, identifier)
	app, err := p.client.Application(orchestratorAppEndpoint)

	orchestratorPolicyEndpoint := fmt.Sprintf("%v/applications/%s/policies", p.orchestratorUrl, identifier)
	foundPolicies, rawJson, anotherPossibleErr := p.client.GetPolicies(orchestratorPolicyEndpoint)

	if err != nil || anotherPossibleErr != nil {
		model := websupport.Model{Map: map[string]interface{}{"resource": "applications", "message": "Error communicating with the orchestrator."}}
		_ = websupport.ModelAndView(w, "applications_show", model)
		log.Println(err)
		return
	}

	var buffer bytes.Buffer
	_ = json.Indent(&buffer, []byte(rawJson), "", "  ")
	resourceLink := fmt.Sprintf("/applications/%v", identifier)
	model := websupport.Model{Map: map[string]interface{}{"resource": "applications", "resource_link": resourceLink,
		"application": app, "policies": foundPolicies, "rawJson": buffer.String()}}
	_ = websupport.ModelAndView(w, "applications_show", model)
}

func (p appsHandler) Edit(w http.ResponseWriter, r *http.Request) {
	identifier := mux.Vars(r)["id"]
	orchestratorAppsEndpoint := fmt.Sprintf("%v/applications/%s", p.orchestratorUrl, identifier)
	app, err := p.client.Application(orchestratorAppsEndpoint)

	orchestratorPolicyEndpoint := fmt.Sprintf("%v/applications/%s/policies", p.orchestratorUrl, identifier)
	foundPolicies, rawJson, anotherPossibleErr := p.client.GetPolicies(orchestratorPolicyEndpoint)

	if err != nil || anotherPossibleErr != nil {
		model := websupport.Model{Map: map[string]interface{}{"resource": "applications", "application": app,
			"message": "Error communicating with the orchestrator."}}
		_ = websupport.ModelAndView(w, "applications_edit", model)
		log.Println(err)
		return
	}

	var buffer bytes.Buffer
	_ = json.Indent(&buffer, []byte(rawJson), "", "  ")
	model := websupport.Model{Map: map[string]interface{}{"resource": "applications", "application": app, "policies": foundPolicies, "rawJson": buffer.String()}}
	_ = websupport.ModelAndView(w, "applications_edit", model)
}

func (p appsHandler) Update(w http.ResponseWriter, r *http.Request) {
	identifier := mux.Vars(r)["id"]
	orchestratorPolicyEndpoint := fmt.Sprintf("%v/applications/%s/policies", p.orchestratorUrl, identifier)
	value := r.FormValue("policy")
	err := p.client.SetPolicies(orchestratorPolicyEndpoint, value)

	if err != nil {
		orchestratorAppsEndpoint := fmt.Sprintf("%v/applications/%s", p.orchestratorUrl, identifier)
		app, _ := p.client.Application(orchestratorAppsEndpoint)
		model := websupport.Model{Map: map[string]interface{}{"resource": "applications", "application": app, "message": "Error communicating with the orchestrator."}}
		_ = websupport.ModelAndView(w, "applications_edit", model)
		log.Println(err)
		return
	}
	applicationsEndpoint := fmt.Sprintf("/applications/%v", identifier)
	http.Redirect(w, r, applicationsEndpoint, http.StatusMovedPermanently)
}
