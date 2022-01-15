package admin

import (
	"fmt"
	"github.com/hexa-org/policy-orchestrator/pkg/websupport"
	"net/http"
)

type Status struct {
	URL    string
	Status string
}

type statusHandler struct {
	orchestratorUrl string
	client          Client
}

func NewStatusHandler(orchestratorUrl string, client Client) statusHandler {
	return statusHandler{orchestratorUrl, client}
}

func (p statusHandler) StatusHandler(w http.ResponseWriter, r *http.Request) {
	url := fmt.Sprintf("%v/health", p.orchestratorUrl)
	health, _ := p.client.Health(url)
	status := Status{url, health}
	model := websupport.Model{Map: map[string]interface{}{"resource": "status", "status": status}}
	_ = websupport.ModelAndView(w, "status", model)
}
