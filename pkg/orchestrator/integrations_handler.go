package orchestrator

import (
	"encoding/json"
	"github.com/gorilla/mux"
	"log"
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

func (handler IntegrationsHandler) List(w http.ResponseWriter, r *http.Request) {
	records, err := handler.gateway.Find()
	if err != nil {
		log.Println(err)
	}
	var list Integrations
	for _, rec := range records {
		list.Integrations = append(list.Integrations, Integration{rec.ID, rec.Name, rec.Provider, rec.Key})
	}
	data, _ := json.Marshal(list)
	w.Header().Set("content-type", "application/json")
	w.WriteHeader(http.StatusOK)
	_, err = w.Write(data)
}

func (handler IntegrationsHandler) Create(w http.ResponseWriter, r *http.Request) {
	var jsonRequest Integration
	_ = json.NewDecoder(r.Body).Decode(&jsonRequest)
	_, err := handler.gateway.Create(jsonRequest.Name, jsonRequest.Provider, jsonRequest.Key)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	find, err := handler.gateway.Find()
	_ = handler.worker.Run(find)
	w.WriteHeader(http.StatusCreated)
}

func (handler IntegrationsHandler) Delete(w http.ResponseWriter, r *http.Request) {
	err := handler.gateway.Delete(mux.Vars(r)["id"])
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	w.WriteHeader(http.StatusOK)
}
