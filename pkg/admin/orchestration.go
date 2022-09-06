package admin

import (
	"fmt"
	"log"
	"net/http"

	"github.com/hexa-org/policy-orchestrator/pkg/websupport"
)

type OrchestrationHandler interface {
	New(w http.ResponseWriter, r *http.Request)
	Update(w http.ResponseWriter, r *http.Request)
}

type orchestrationHandler struct {
	orchestratorUrl string
	client          Client
}

func NewOrchestrationHandler(orchestratorUrl string, client Client) OrchestrationHandler {
	return orchestrationHandler{orchestratorUrl, client}
}

func (p orchestrationHandler) New(w http.ResponseWriter, _ *http.Request) {
	foundApplications, clientErr := p.client.Applications()
	if clientErr != nil {
		model := websupport.Model{Map: map[string]interface{}{"resource": "orchestration", "message": clientErr.Error()}}
		_ = websupport.ModelAndView(w, &resources, "orchestration_new", model)
		log.Println(clientErr)
		return
	}
	model := websupport.Model{Map: map[string]interface{}{"resource": "orchestration", "applications": foundApplications}}
	_ = websupport.ModelAndView(w, &resources, "orchestration_new", model)
}

func (p orchestrationHandler) Update(w http.ResponseWriter, r *http.Request) {
	err := Apply(r.FormValue("from"), r.FormValue("to"))
	if err != nil {
		fmt.Println(Apply(r.FormValue("from"), r.FormValue("to")).Error())
	}
	http.Redirect(w, r, "/applications", http.StatusMovedPermanently)
}

func Apply(from, to string) error {
	fmt.Println(from)
	fmt.Println(to)
	return nil
}
