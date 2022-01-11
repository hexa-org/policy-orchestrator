package orchestrator

import (
	"encoding/json"
	"net/http"
)

type Applications struct {
	Applications []Application `json:"applications"`
}

type Application struct {
	ID          string `json:"id"`
	ObjectId    string `json:"object_id"`
	Name        string `json:"name"`
	Description string `json:"description"`
}

type ApplicationsHandler struct {
	gateway ApplicationsDataGateway
}

func (handler ApplicationsHandler) List(w http.ResponseWriter, r *http.Request) {
	records, _ := handler.gateway.Find()
	var list Applications
	for _, rec := range records {
		list.Applications = append(list.Applications, Application{rec.ID, rec.ObjectId, rec.Name, rec.Description})
	}
	data, _ := json.Marshal(list)
	w.Header().Set("content-type", "application/json")
	w.WriteHeader(http.StatusOK)
	_, _ = w.Write(data)
}
