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

type StatusHandler struct {
	orchestratorUrl string
	client          Client
}

func NewStatusHandler(orchestratorUrl string, client Client) StatusHandler {
	return StatusHandler{orchestratorUrl, client}
}

func (p StatusHandler) StatusHandler(w http.ResponseWriter, _ *http.Request) {
	url := fmt.Sprintf("%v/health", p.orchestratorUrl)
	health, _ := p.client.Health(url)
	status := Status{url, health}
	model := websupport.Model{Map: map[string]interface{}{"resource": "status", "status": status}}
	_ = websupport.ModelAndView(w, "status", model)
}
