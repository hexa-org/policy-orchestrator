package orchestrator

import (
	"encoding/json"
	"log"
	"net/http"
	"sort"
	"strings"

	"github.com/hexa-org/policy-mapper/pkg/hexapolicy"
	"github.com/hexa-org/policy-orchestrator/demo/pkg/dataConfigGateway"

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

type ApplicationsHandler struct {
	applicationsGateway dataConfigGateway.ApplicationsDataGateway
	integrationsGateway dataConfigGateway.IntegrationsDataGateway
	applicationsService ApplicationsService
}

func (handler ApplicationsHandler) List(w http.ResponseWriter, r *http.Request) {
	doRefresh := false
	refresh := r.URL.Query().Get("refresh")
	if refresh == "true" {
		doRefresh = true
	}
	integrationRecords := handler.integrationsGateway.Find()

	integrationNamesById := make(map[string]string)
	for _, integration := range integrationRecords {
		integrationNamesById[integration.ID] = integration.Provider // todo - include provider within applications table
	}

	records, applicationErr := handler.applicationsGateway.Find(doRefresh)
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

	policies := hexapolicy.Policies{
		Policies: records,
		App:      &application.ObjectID,
	}

	data, _ := json.Marshal(policies)
	w.Header().Set("content-type", "application/json")
	w.WriteHeader(http.StatusOK)
	_, _ = w.Write(data)
}

func (handler ApplicationsHandler) SetPolicies(w http.ResponseWriter, r *http.Request) {
	var policies hexapolicy.Policies
	if erroneousDecode := json.NewDecoder(r.Body).Decode(&policies); erroneousDecode != nil {
		http.Error(w, erroneousDecode.Error(), http.StatusInternalServerError)
		return
	}

	validatorErr := validator.New().Var(policies.Policies, "omitempty,dive")
	if validatorErr != nil {
		http.Error(w, "unable to validate policy.", http.StatusInternalServerError)
		return
	}

	application, integration, provider, err := handler.applicationsService.GatherRecords(mux.Vars(r)["id"])
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	status, setErr := provider.SetPolicyInfo(integration, application, policies.Policies)
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
