package admin

import (
	"bytes"
	"encoding/json"
	"fmt"
	"log"
	"net/http"

	"github.com/gorilla/mux"
	"github.com/hexa-org/policy-mapper/pkg/sessionSupport"
	"github.com/hexa-org/policy-mapper/pkg/websupport"
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
	session         sessionSupport.SessionManager
}

func NewApplicationsHandler(sessionHandler sessionSupport.SessionManager, orchestratorUrl string, client Client) ApplicationsHandler {
	return appsHandler{orchestratorUrl, client, sessionHandler}
}

func (p appsHandler) List(w http.ResponseWriter, r *http.Request) {
	doRefresh := false
	refresh := r.URL.Query().Get("refresh")
	if refresh == "true" {
		doRefresh = true
	}
	foundApplications, clientErr := p.client.Applications(doRefresh)
	if clientErr != nil {
		model := websupport.Model{Map: map[string]interface{}{"resource": "applications", "message": clientErr.Error()}}
		_ = websupport.ModelAndView(w, &resources, "applications", model)
		log.Println(clientErr)
		return
	}
	sessionInfo, err := p.session.Session(r)
	if err != nil {
		sessionInfo = &sessionSupport.SessionInfo{}
	}
	model := websupport.Model{Map: map[string]interface{}{"resource": "applications", "applications": foundApplications, "session": sessionInfo}}
	_ = websupport.ModelAndView(w, &resources, "applications", model)
	w.Header()
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
	sessionInfo, err := p.session.Session(r)
	if err != nil {
		sessionInfo = &sessionSupport.SessionInfo{}
	}
	model := websupport.Model{Map: map[string]interface{}{"resource": "applications", "resource_link": resourceLink, "application": foundApplication, "policies": foundPolicies, "rawJson": buffer.String(), "session": sessionInfo}}
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

	sessionInfo, err := p.session.Session(r)
	if err != nil {
		sessionInfo = &sessionSupport.SessionInfo{}
	}
	foundPolicies, rawJson, policiesError := p.client.GetPolicies(identifier)
	if policiesError != nil {

		model := websupport.Model{Map: map[string]interface{}{"resource": "applications", "application": foundApplication, "message": policiesError.Error(), "session": sessionInfo}}
		_ = websupport.ModelAndView(w, &resources, "applications_edit", model)
		log.Println(policiesError)
		return
	}

	var buffer bytes.Buffer
	_ = json.Indent(&buffer, []byte(rawJson), "", "  ")
	model := websupport.Model{Map: map[string]interface{}{"resource": "applications", "application": foundApplication, "policies": foundPolicies, "rawJson": buffer.String(), "session": sessionInfo}}
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
