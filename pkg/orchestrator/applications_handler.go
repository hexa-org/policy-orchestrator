package orchestrator

import (
	"encoding/json"
	"net/http"
)

type Applications struct {
	Applications []Application `json:"applications"`
}

type Application struct {
	Name string `json:"name"`
}

type ApplicationsHandler struct{}

func (handler ApplicationsHandler) List(w http.ResponseWriter, r *http.Request) {
	data, _ := json.Marshal(Applications{[]Application{{"anApp"}}})
	w.Header().Set("content-type", "application/json")
	w.WriteHeader(http.StatusOK)
	_, _ = w.Write(data)
}
