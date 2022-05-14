package orchestrator

import (
	"encoding/json"
	"github.com/gorilla/mux"
	"github.com/hexa-org/policy-orchestrator/pkg/policysupport"
	"log"
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
	providers           map[string]Provider
}

func (handler ApplicationsHandler) List(w http.ResponseWriter, _ *http.Request) {
	records, err := handler.applicationsGateway.Find()
	if err != nil {
		log.Println("Error accessing database: " + err.Error())
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
	applicationRecord, integrationRecord, err := handler.gatherRecords(r)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	integration := IntegrationInfo{Name: integrationRecord.Name, Key: integrationRecord.Key}
	application := ApplicationInfo{ObjectID: applicationRecord.ObjectId, Name: applicationRecord.Name, Description: applicationRecord.Description}
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

func (handler ApplicationsHandler) SetPolicies(w http.ResponseWriter, r *http.Request) {
	applicationRecord, integrationRecord, err := handler.gatherRecords(r)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	var policies []Policy
	if erroneousDecode := json.NewDecoder(r.Body).Decode(&policies); erroneousDecode != nil {
		http.Error(w, erroneousDecode.Error(), http.StatusInternalServerError)
		return
	}

	integration := IntegrationInfo{Name: integrationRecord.Name, Key: integrationRecord.Key}
	application := ApplicationInfo{ObjectID: applicationRecord.ObjectId, Name: applicationRecord.Name, Description: applicationRecord.Description}
	pro := handler.providers[strings.ToLower(integrationRecord.Provider)] // todo - test for lower?
	var policyInfos []policysupport.PolicyInfo
	for _, policy := range policies {
		info := policysupport.PolicyInfo{Version: policy.Version, Action: policy.Action,
			Subject: policysupport.SubjectInfo{AuthenticatedUsers: policy.Subject.AuthenticatedUsers},
			Object:  policysupport.ObjectInfo{Resources: policy.Object.Resources}}
		policyInfos = append(policyInfos, info)
	}
	status, setErr := pro.SetPolicyInfo(integration, application, policyInfos)
	if setErr != nil || status != 201 {
		http.Error(w, "unable to update policy.", http.StatusInternalServerError)
		return
	}
	w.WriteHeader(status)
}

func (handler ApplicationsHandler) gatherRecords(r *http.Request) (ApplicationRecord, IntegrationRecord, error) {
	identifier := mux.Vars(r)["id"]
	applicationRecord, err := handler.applicationsGateway.FindById(identifier)
	if err != nil {
		return ApplicationRecord{}, IntegrationRecord{}, err
	}
	integrationRecord, err := handler.integrationsGateway.FindById(applicationRecord.IntegrationId)
	if err != nil {
		return ApplicationRecord{}, IntegrationRecord{}, err
	}
	return applicationRecord, integrationRecord, err
}
