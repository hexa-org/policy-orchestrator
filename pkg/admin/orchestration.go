package admin

import (
	"fmt"
	"log"
	"net/http"

	"github.com/hexa-org/policy-orchestrator/pkg/websupport"
)

type OrchestrationHandler interface {
	New(w http.ResponseWriter, r *http.Request)
	Apply(w http.ResponseWriter, r *http.Request)
}

type orchestrationHandler struct {
	orchestratorUrl string
	client          Client
}

func NewOrchestrationHandler(orchestratorUrl string, client Client) OrchestrationHandler {
	return orchestrationHandler{orchestratorUrl, client}
}

func (p orchestrationHandler) New(w http.ResponseWriter, _ *http.Request) {
	orchestratorEndpoint := fmt.Sprintf("%v/applications", p.orchestratorUrl)
	foundApplications, clientErr := p.client.Applications(orchestratorEndpoint)
	if clientErr != nil {
		model := websupport.Model{Map: map[string]interface{}{"resource": "orchestration", "message": clientErr.Error()}}
		_ = websupport.ModelAndView(w, &resources, "orchestration_new", model)
		log.Println(clientErr)
		return
	}
	model := websupport.Model{Map: map[string]interface{}{"resource": "orchestration", "applications": foundApplications}}
	_ = websupport.ModelAndView(w, &resources, "orchestration_new", model)
}

func (p orchestrationHandler) Apply(w http.ResponseWriter, r *http.Request) {
	http.Redirect(w, r, "/applications", http.StatusMovedPermanently)
}
