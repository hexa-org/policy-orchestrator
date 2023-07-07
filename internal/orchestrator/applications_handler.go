package orchestrator

import (
	"encoding/json"
	"log"
	"net/http"
	"sort"
	"strings"

	"github.com/hexa-org/policy-orchestrator/internal/policysupport"

	"github.com/go-playground/validator/v10"
	"github.com/gorilla/mux"
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
	Service       string `json:"service"`
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
	applicationsService ApplicationsService
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
		integrationNamesById[integration.ID] = integration.Provider // todo - include provider within applications table
	}

	records, applicationErr := handler.applicationsGateway.Find()
	if applicationErr != nil {
		log.Println("Error accessing database: " + applicationErr.Error())
		http.Error(w, applicationErr.Error(), http.StatusInternalServerError)
		return
	}

	var list Applications
	for _, rec := range records {
		list.Applications = append(list.Applications, Application{ID: rec.ID, IntegrationId: rec.IntegrationId, ObjectId: rec.ObjectId, Name: rec.Name, Description: rec.Description, ProviderName: integrationNamesById[rec.IntegrationId], Service: rec.Service})
	}

	// sort by "Provider" so that all app resources from a platform are grouped together.
	// sort ProviderName as the first (asc) order and Service as the second (asc) order.
	sort.Slice(list.Applications, func(i, j int) bool {
		providerComp := strings.Compare(list.Applications[i].ProviderName, list.Applications[j].ProviderName)
		serviceComp := strings.Compare(list.Applications[i].Service, list.Applications[j].Service)
		switch providerComp {
		case 0:
			return serviceComp <= 0
		default:
			return providerComp < 0
		}
	})

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
	app := Application{
		ID:            record.ID,
		IntegrationId: record.IntegrationId,
		ObjectId:      record.ObjectId,
		Name:          record.Name,
		Description:   record.Description,
		Service:       record.Service,
	}
	data, _ := json.Marshal(app)
	w.Header().Set("content-type", "application/json")
	w.WriteHeader(http.StatusOK)
	_, _ = w.Write(data)
}

func (handler ApplicationsHandler) GetPolicies(w http.ResponseWriter, r *http.Request) {
	application, integration, provider, err := handler.applicationsService.GatherRecords(mux.Vars(r)["id"])
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	records, err := provider.GetPolicyInfo(integration, application)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	list := make([]Policy, 0)
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
	var policies Policies
	if erroneousDecode := json.NewDecoder(r.Body).Decode(&policies); erroneousDecode != nil {
		http.Error(w, erroneousDecode.Error(), http.StatusInternalServerError)
		return
	}

	validatorErr := validator.New().Var(policies.Policies, "omitempty,dive")
	if validatorErr != nil {
		http.Error(w, "unable to validate policy.", http.StatusInternalServerError)
		return
	}

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

	application, integration, provider, err := handler.applicationsService.GatherRecords(mux.Vars(r)["id"])
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	status, setErr := provider.SetPolicyInfo(integration, application, policyInfos)
	if setErr != nil {
		log.Printf("unable to update policy: %s", setErr.Error())
		// todo - should we return the error msg here.
		http.Error(w, "unable to update policy.", http.StatusInternalServerError)
		return
	}

	if status != http.StatusCreated {
		http.Error(w, "unable to update policy.", http.StatusInternalServerError)
		return
	}

	w.WriteHeader(status)
}
