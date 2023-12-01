package admin

import (
	"bytes"
	"encoding/json"
	"fmt"
	"log"
	"net/http"

	"github.com/gorilla/mux"
	"github.com/hexa-org/policy-orchestrator/demo/pkg/websupport"
)

type Application struct {
	ID            string
	IntegrationId string
	ObjectId      string
	Name          string
	Description   string
	ProviderName  string
	Service       string
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
	Policies(w http.ResponseWriter, r *http.Request)
}

type appsHandler struct {
	orchestratorUrl string
	client          Client
}

func NewApplicationsHandler(orchestratorUrl string, client Client) ApplicationsHandler {
	return appsHandler{orchestratorUrl, client}
}

func (p appsHandler) List(w http.ResponseWriter, _ *http.Request) {
	foundApplications, clientErr := p.client.Applications()
	if clientErr != nil {
		model := websupport.Model{Map: map[string]interface{}{"resource": "applications", "message": clientErr.Error()}}
		_ = websupport.ModelAndView(w, &resources, "applications", model)
		log.Println(clientErr)
		return
	}
	model := websupport.Model{Map: map[string]interface{}{"resource": "applications", "applications": foundApplications}}
	_ = websupport.ModelAndView(w, &resources, "applications", model)
}

func (p appsHandler) Show(w http.ResponseWriter, r *http.Request) {
	identifier := mux.Vars(r)["id"]

	foundApplication, clientErr := p.client.Application(identifier)
	if clientErr != nil {
		model := websupport.Model{Map: map[string]interface{}{"resource": "applications", "message": clientErr.Error()}}
		_ = websupport.ModelAndView(w, &resources, "applications_show", model)
		log.Println(clientErr)
		return
	}

	// / todo - consider one rest call here?
	foundPolicies, rawJson, policiesError := p.client.GetPolicies(identifier)
	if policiesError != nil {
		model := websupport.Model{Map: map[string]interface{}{"resource": "applications", "message": policiesError.Error()}}
		_ = websupport.ModelAndView(w, &resources, "applications_show", model)
		log.Println(policiesError)
		return
	}

	var buffer bytes.Buffer
	_ = json.Indent(&buffer, []byte(rawJson), "", "  ")
	resourceLink := fmt.Sprintf("/applications/%v", identifier)
	model := websupport.Model{Map: map[string]interface{}{"resource": "applications", "resource_link": resourceLink, "application": foundApplication, "policies": foundPolicies, "rawJson": buffer.String()}}
	_ = websupport.ModelAndView(w, &resources, "applications_show", model)
}

func (p appsHandler) Edit(w http.ResponseWriter, r *http.Request) {
	identifier := mux.Vars(r)["id"]

	foundApplication, applicationError := p.client.Application(identifier)
	if applicationError != nil {
		model := websupport.Model{Map: map[string]interface{}{"resource": "applications", "application": foundApplication, "message": applicationError.Error()}}
		_ = websupport.ModelAndView(w, &resources, "applications_edit", model)
		log.Println(applicationError)
		return
	}

	foundPolicies, rawJson, policiesError := p.client.GetPolicies(identifier)
	if policiesError != nil {
		model := websupport.Model{Map: map[string]interface{}{"resource": "applications", "application": foundApplication, "message": policiesError.Error()}}
		_ = websupport.ModelAndView(w, &resources, "applications_edit", model)
		log.Println(policiesError)
		return
	}

	var buffer bytes.Buffer
	_ = json.Indent(&buffer, []byte(rawJson), "", "  ")
	model := websupport.Model{Map: map[string]interface{}{"resource": "applications", "application": foundApplication, "policies": foundPolicies, "rawJson": buffer.String()}}
	_ = websupport.ModelAndView(w, &resources, "applications_edit", model)
}

func (p appsHandler) Update(w http.ResponseWriter, r *http.Request) {
	identifier := mux.Vars(r)["id"]
	desiredPolicies := r.FormValue("policy")
	clientErr := p.client.SetPolicies(identifier, desiredPolicies)

	if clientErr != nil {
		foundApplication, _ := p.client.Application(identifier)
		model := websupport.Model{Map: map[string]interface{}{"resource": "applications", "application": foundApplication, "policies": desiredPolicies, "message": clientErr.Error()}}
		_ = websupport.ModelAndView(w, &resources, "applications_edit", model)
		log.Println(clientErr)
		return
	}
	applicationsEndpoint := fmt.Sprintf("/applications/%v", identifier)
	http.Redirect(w, r, applicationsEndpoint, http.StatusMovedPermanently)
}

// todo - maybe the below should be closer to the orchestration functions

func (p appsHandler) Policies(w http.ResponseWriter, r *http.Request) {
	identifier := mux.Vars(r)["id"]
	_, rawJson, _ := p.client.GetPolicies(identifier)
	var buffer bytes.Buffer
	_ = json.Indent(&buffer, []byte(rawJson), "", "  ")
	w.Header().Set("content-type", "application/json")
	w.WriteHeader(http.StatusOK)
	_, _ = w.Write(buffer.Bytes())
}
