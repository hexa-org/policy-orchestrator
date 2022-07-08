package orchestrator

import (
	"encoding/json"
	"github.com/go-playground/validator/v10"
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
	ProviderName  string `json:"provider_name"`
}

type Policies struct {
	Policies []Policy `json:"policies"`
}

type Policy struct {
	Meta    Meta     `json:"meta" validate:"required"`
	Actions []Action `json:"actions" validate:"required"`
	Subject Subject  `json:"subject" validate:"required"`
	Object  Object   `json:"object" validate:"required"`
}

type Meta struct {
	Version string `json:"version" validate:"required"`
}

type Action struct {
	ActionUri string `json:"action_uri" validate:"required"`
}

type Subject struct {
	Members []string `json:"members" validate:"required"`
}

type Object struct {
	ResourceID string `json:"resource_id" validate:"required"`
}

type ApplicationsHandler struct {
	applicationsGateway ApplicationsDataGateway
	integrationsGateway IntegrationsDataGateway
	providers           map[string]Provider
}

func (handler ApplicationsHandler) List(w http.ResponseWriter, _ *http.Request) {
	integrationRecords, integrationErr := handler.integrationsGateway.Find()
	if integrationErr != nil {
		log.Println("Error accessing database: " + integrationErr.Error())
		http.Error(w, integrationErr.Error(), http.StatusInternalServerError)
		return
	}

	integrationNamesById := make(map[string]string, 0)
	for _, integration := range integrationRecords {
		integrationNamesById[integration.ID] = integration.Provider
	}

	records, applicationErr := handler.applicationsGateway.Find()
	if applicationErr != nil {
		log.Println("Error accessing database: " + applicationErr.Error())
		http.Error(w, applicationErr.Error(), http.StatusInternalServerError)
		return
	}

	var list Applications
	for _, rec := range records {
		list.Applications = append(list.Applications, Application{ID: rec.ID, IntegrationId: rec.IntegrationId, ObjectId: rec.ObjectId, Name: rec.Name, Description: rec.Description, ProviderName: integrationNamesById[rec.IntegrationId]})
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
	app := Application{ID: record.ID, IntegrationId: record.IntegrationId, ObjectId: record.ObjectId, Name: record.Name, Description: record.Description}
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
		var actions []Action
		for _, a := range rec.Actions {
			actions = append(actions, Action{a.ActionUri})
		}
		list = append(
			list, Policy{
				Meta: Meta{
					rec.Meta.Version,
				},
				Actions: actions,
				Subject: Subject{
					rec.Subject.Members,
				},
				Object: Object{
					ResourceID: rec.Object.ResourceID,
				},
			})
	}
	data, _ := json.Marshal(Policies{list})
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

	var policies Policies
	if erroneousDecode := json.NewDecoder(r.Body).Decode(&policies); erroneousDecode != nil {
		http.Error(w, erroneousDecode.Error(), http.StatusInternalServerError)
		return
	}

	err = validator.New().Var(policies.Policies, "omitempty,dive")
	if err != nil {
		http.Error(w, "unable to validate policy.", http.StatusInternalServerError)
		return
	}

	integration := IntegrationInfo{Name: integrationRecord.Name, Key: integrationRecord.Key}
	application := ApplicationInfo{ObjectID: applicationRecord.ObjectId, Name: applicationRecord.Name, Description: applicationRecord.Description}
	pro := handler.providers[strings.ToLower(integrationRecord.Provider)] // todo - test for lower?
	var policyInfos []policysupport.PolicyInfo
	for _, policy := range policies.Policies {
		var actionInfos []policysupport.ActionInfo
		for _, a := range policy.Actions {
			actionInfos = append(actionInfos, policysupport.ActionInfo{ActionUri: a.ActionUri})
		}
		info := policysupport.PolicyInfo{
			Meta: policysupport.MetaInfo{
				Version: policy.Meta.Version,
			},
			Actions: actionInfos,
			Subject: policysupport.SubjectInfo{
				Members: policy.Subject.Members,
			},
			Object: policysupport.ObjectInfo{
				ResourceID: policy.Object.ResourceID,
			},
		}
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
