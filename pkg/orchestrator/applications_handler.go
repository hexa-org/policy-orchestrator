package orchestrator

import (
	"encoding/json"
	"github.com/gorilla/mux"
	"github.com/hexa-org/policy-orchestrator/pkg/orchestrator/provider"
	"net/http"
	"strings"
)

type Applications struct {
	Applications []Application `json:"applications"`
}

type Application struct {
	ID            string `json:"id"`
	IntegrationId string `json:"integration_id"`
	ObjectId      string `json:"object_id"`
	Name          string `json:"name"`
	Description   string `json:"description"`
}

type Policy struct {
	Version string  `json:"version"`
	Action  string  `json:"action"`
	Subject Subject `json:"subject"`
	Object  Object  `json:"object"`
}

type Subject struct {
	AuthenticatedUsers []string `json:"authenticated_users"`
}

type Object struct {
	Resources []string `json:"resources"`
}

type ApplicationsHandler struct {
	applicationsGateway ApplicationsDataGateway
	integrationsGateway IntegrationsDataGateway
	providers           map[string]provider.Provider
}

func (handler ApplicationsHandler) List(w http.ResponseWriter, r *http.Request) {
	records, err := handler.applicationsGateway.Find()
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	var list Applications
	for _, rec := range records {
		list.Applications = append(list.Applications, Application{rec.ID, rec.IntegrationId, rec.ObjectId, rec.Name, rec.Description})
	}
	data, _ := json.Marshal(list)
	w.Header().Set("content-type", "application/json")
	w.WriteHeader(http.StatusOK)
	_, _ = w.Write(data)
}

func (handler ApplicationsHandler) Show(w http.ResponseWriter, r *http.Request) {
	identifier := mux.Vars(r)["id"]
	record, err := handler.applicationsGateway.FindById(identifier)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	app := Application{record.ID, record.IntegrationId, record.ObjectId, record.Name, record.Description}
	data, _ := json.Marshal(app)
	w.Header().Set("content-type", "application/json")
	w.WriteHeader(http.StatusOK)
	_, _ = w.Write(data)
}

func (handler ApplicationsHandler) GetPolicies(w http.ResponseWriter, r *http.Request) {
	identifier := mux.Vars(r)["id"]
	record, err := handler.applicationsGateway.FindById(identifier)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	integrationRecord, err := handler.integrationsGateway.FindById(record.IntegrationId)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	integration := provider.IntegrationInfo{Name: integrationRecord.Name, Key: integrationRecord.Key}
	application := provider.ApplicationInfo{ObjectID: record.ObjectId, Name: record.Name, Description: record.Description}
	p := handler.providers[strings.ToLower(integrationRecord.Provider)] // todo - test for lower?
	records, err := p.GetPolicyInfo(integration, application)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	var list []Policy
	for _, rec := range records {
		list = append(
			list, Policy{
				rec.Version,
				rec.Action,
				Subject{rec.Subject.AuthenticatedUsers},
				Object{rec.Object.Resources},
			})
	}
	data, _ := json.Marshal(list)
	w.Header().Set("content-type", "application/json")
	w.WriteHeader(http.StatusOK)
	_, _ = w.Write(data)
}
