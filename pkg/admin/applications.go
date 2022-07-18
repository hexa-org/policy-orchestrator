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
	orchestratorEndpoint := fmt.Sprintf("%v/applications", p.orchestratorUrl)
	foundApplications, clientErr := p.client.Applications(orchestratorEndpoint)
	if clientErr != nil {
		model := websupport.Model{Map: map[string]interface{}{"resource": "applications", "message": clientErr.Error()}}
		_ = websupport.ModelAndView(w, "applications", model)
		log.Println(clientErr)
		return
	}
	model := websupport.Model{Map: map[string]interface{}{"resource": "applications", "applications": foundApplications}}
	_ = websupport.ModelAndView(w, "applications", model)
}

func (p appsHandler) Show(w http.ResponseWriter, r *http.Request) {
	identifier := mux.Vars(r)["id"]
	orchestratorEndpoint := fmt.Sprintf("%v/applications/%s", p.orchestratorUrl, identifier)
	foundApplication, clientErr := p.client.Application(orchestratorEndpoint)
	if clientErr != nil {
		model := websupport.Model{Map: map[string]interface{}{"resource": "applications", "message": clientErr.Error()}}
		_ = websupport.ModelAndView(w, "applications_show", model)
		log.Println(clientErr)
		return
	}

	/// todo - consider one rest call here?
	orchestratorPoliciesEndpoint := fmt.Sprintf("%v/applications/%s/policies", p.orchestratorUrl, identifier)
	foundPolicies, rawJson, policiesError := p.client.GetPolicies(orchestratorPoliciesEndpoint)
	if policiesError != nil {
		model := websupport.Model{Map: map[string]interface{}{"resource": "applications", "message": policiesError.Error()}}
		_ = websupport.ModelAndView(w, "applications_show", model)
		log.Println(policiesError)
		return
	}

	var buffer bytes.Buffer
	_ = json.Indent(&buffer, []byte(rawJson), "", "  ")
	resourceLink := fmt.Sprintf("/applications/%v", identifier)
	model := websupport.Model{Map: map[string]interface{}{"resource": "applications", "resource_link": resourceLink, "application": foundApplication, "policies": foundPolicies, "rawJson": buffer.String()}}
	_ = websupport.ModelAndView(w, "applications_show", model)
}

func (p appsHandler) Edit(w http.ResponseWriter, r *http.Request) {
	identifier := mux.Vars(r)["id"]
	orchestratorEndpoint := fmt.Sprintf("%v/applications/%s", p.orchestratorUrl, identifier)
	foundApplication, applicationError := p.client.Application(orchestratorEndpoint)

	if applicationError != nil {
		model := websupport.Model{Map: map[string]interface{}{"resource": "applications", "application": foundApplication, "message": applicationError.Error()}}
		_ = websupport.ModelAndView(w, "applications_edit", model)
		log.Println(applicationError)
		return
	}

	orchestratorPoliciesEndpoint := fmt.Sprintf("%v/applications/%s/policies", p.orchestratorUrl, identifier)
	foundPolicies, rawJson, policiesError := p.client.GetPolicies(orchestratorPoliciesEndpoint)
	if policiesError != nil {
		model := websupport.Model{Map: map[string]interface{}{"resource": "applications", "application": foundApplication, "message": policiesError.Error()}}
		_ = websupport.ModelAndView(w, "applications_edit", model)
		log.Println(policiesError)
		return
	}

	var buffer bytes.Buffer
	_ = json.Indent(&buffer, []byte(rawJson), "", "  ")
	model := websupport.Model{Map: map[string]interface{}{"resource": "applications", "application": foundApplication, "policies": foundPolicies, "rawJson": buffer.String()}}
	_ = websupport.ModelAndView(w, "applications_edit", model)
}

func (p appsHandler) Update(w http.ResponseWriter, r *http.Request) {
	identifier := mux.Vars(r)["id"]
	orchestratorEndpoint := fmt.Sprintf("%v/applications/%s/policies", p.orchestratorUrl, identifier)
	desiredPolicies := r.FormValue("policy")
	clientErr := p.client.SetPolicies(orchestratorEndpoint, desiredPolicies)

	if clientErr != nil {
		orchestratorAppsEndpoint := fmt.Sprintf("%v/applications/%s", p.orchestratorUrl, identifier)
		foundApplication, _ := p.client.Application(orchestratorAppsEndpoint)
		model := websupport.Model{Map: map[string]interface{}{"resource": "applications", "application": foundApplication, "policies": desiredPolicies, "message": clientErr.Error()}}
		_ = websupport.ModelAndView(w, "applications_edit", model)
		log.Println(clientErr)
		return
	}
	applicationsEndpoint := fmt.Sprintf("/applications/%v", identifier)
	http.Redirect(w, r, applicationsEndpoint, http.StatusMovedPermanently)
}
