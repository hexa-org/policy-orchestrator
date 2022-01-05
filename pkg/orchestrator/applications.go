package orchestrator

import (
	"encoding/json"
	"net/http"
)

type applications struct {
	Applications []application `json:"applications"`
}

type application struct {
	Name string `json:"name"`
}

type handler struct{}

func NewApplicationsHandler() handler { return handler{} }

func (a handler) Applications(w http.ResponseWriter, r *http.Request) {
	data, _ := json.Marshal(applications{[]application{{"anApp"}}})
	w.Header().Set("content-type", "application/json")
	w.WriteHeader(http.StatusOK)
	_, _ = w.Write(data)
}
