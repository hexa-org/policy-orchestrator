package orchestrator

import (
	"encoding/json"
	"github.com/gorilla/mux"
	"log"
	"net/http"
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

type ApplicationsHandler struct {
	gateway ApplicationsDataGateway
}

func (handler ApplicationsHandler) List(w http.ResponseWriter, r *http.Request) {
	records, err := handler.gateway.Find()
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	var list Applications
	for _, rec := range records {
		list.Applications = append(list.Applications, Application{rec.ID, rec.IntegreationId, rec.ObjectId, rec.Name, rec.Description})
	}
	data, _ := json.Marshal(list)
	w.Header().Set("content-type", "application/json")
	w.WriteHeader(http.StatusOK)
	_, err = w.Write(data)
	if err != nil {
		log.Println(err)
	}
}

func (handler ApplicationsHandler) Show(w http.ResponseWriter, r *http.Request) {
	identifier := mux.Vars(r)["id"]
	record, err := handler.gateway.FindById(identifier)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	app := Application{record.ID, record.IntegreationId, record.ObjectId, record.Name, record.Description}
	data, _ := json.Marshal(app)
	w.Header().Set("content-type", "application/json")
	w.WriteHeader(http.StatusOK)
	_, err = w.Write(data)
	if err != nil {
		log.Println(err)
	}
}
