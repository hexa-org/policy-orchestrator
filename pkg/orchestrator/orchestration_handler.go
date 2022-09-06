package orchestrator

import (
	"net/http"
)

type OrchestrationHandler struct {
}

func (o OrchestrationHandler) Update(writer http.ResponseWriter, _ *http.Request) {
	writer.WriteHeader(http.StatusCreated)
}
