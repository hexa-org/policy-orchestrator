package orchestrator

import (
	"encoding/json"
	"net/http"
)

type OrchestrationHandler struct {
	applicationsService ApplicationsService
}

type Orchestration struct {
	From string `json:"from"`
	To   string `json:"to"`
}

func (o OrchestrationHandler) Update(writer http.ResponseWriter, request *http.Request) {
	var jsonRequest Orchestration
	_ = json.NewDecoder(request.Body).Decode(&jsonRequest)
	err := o.applicationsService.Apply(jsonRequest)
	if err != nil {
		http.Error(writer, err.Error(), http.StatusInternalServerError)
		return
	}
	writer.WriteHeader(http.StatusCreated)
}
