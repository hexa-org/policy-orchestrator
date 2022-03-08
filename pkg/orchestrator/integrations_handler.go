package orchestrator

import (
	"encoding/json"
	"github.com/gorilla/mux"
	"net/http"
)

type Integrations struct {
	Integrations []Integration `json:"integrations"`
}

type Integration struct {
	ID       string `json:"id"`
	Name     string `json:"name"`
	Provider string `json:"provider"`
	Key      []byte `json:"key"`
}

type IntegrationsHandler struct {
	gateway IntegrationsDataGateway
	worker  DiscoveryWorker
}

func (handler IntegrationsHandler) List(w http.ResponseWriter, _ *http.Request) {
	records, _ := handler.gateway.Find()
	var list Integrations
	for _, rec := range records {
		list.Integrations = append(list.Integrations, Integration{rec.ID, rec.Name, rec.Provider, rec.Key})
	}
	data, _ := json.Marshal(list)
	w.Header().Set("content-type", "application/json")
	w.WriteHeader(http.StatusOK)
	_, _ = w.Write(data)
}

func (handler IntegrationsHandler) Create(w http.ResponseWriter, r *http.Request) {
	var jsonRequest Integration
	_ = json.NewDecoder(r.Body).Decode(&jsonRequest)
	_, err := handler.gateway.Create(jsonRequest.Name, jsonRequest.Provider, jsonRequest.Key)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	handler.RunWorker()
	w.WriteHeader(http.StatusCreated)
}

func (handler IntegrationsHandler) Delete(w http.ResponseWriter, r *http.Request) {
	if err := handler.gateway.Delete(mux.Vars(r)["id"]); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	w.WriteHeader(http.StatusOK)
}

func (handler IntegrationsHandler) RunWorker() {
	find, _ := handler.gateway.Find()
	_ = handler.worker.Run(find)
}
