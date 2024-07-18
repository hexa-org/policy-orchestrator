package admin

import (
	"log"
	"net/http"

	"github.com/hexa-org/policy-mapper/pkg/sessionSupport"
	"github.com/hexa-org/policy-mapper/pkg/websupport"
)

type OrchestrationHandler interface {
	New(w http.ResponseWriter, r *http.Request)
	Update(w http.ResponseWriter, r *http.Request)
}

type orchestrationHandler struct {
	orchestratorUrl string
	client          Client
	session         sessionSupport.SessionManager
}

func NewOrchestrationHandler(orchestratorUrl string, client Client, sessionHandler sessionSupport.SessionManager) OrchestrationHandler {
	return orchestrationHandler{orchestratorUrl, client, sessionHandler}
}

func (p orchestrationHandler) New(w http.ResponseWriter, r *http.Request) {
	foundApplications, clientErr := p.client.Applications(false)
	sessionInfo, err := p.session.Session(r)
	if err != nil {
		sessionInfo = &sessionSupport.SessionInfo{}
	}
	if clientErr != nil {

		model := websupport.Model{Map: map[string]interface{}{"resource": "orchestration", "message": clientErr.Error(), "session": sessionInfo}}
		_ = websupport.ModelAndView(w, &resources, "orchestration_new", model)
		log.Println(clientErr)
		return
	}

	// / todo - remove once working across providers
	// TODO - Check this with Gerry
	/*available := make([]Application, 0)
	  for _, app := range foundApplications {
	  	if app.ProviderName == "google_cloud" || app.ProviderName == "open_policy_agent" || app.ProviderName == "azure" {
	  		available = append(available, app)
	  	}
	  }*/

	model := websupport.Model{Map: map[string]interface{}{"resource": "orchestration", "applications": foundApplications, "session": sessionInfo}}
	_ = websupport.ModelAndView(w, &resources, "orchestration_new", model)
}

func (p orchestrationHandler) Update(w http.ResponseWriter, r *http.Request) {
	clientErr := p.client.Orchestration(r.FormValue("from"), r.FormValue("to"))
	if clientErr != nil {
		model := websupport.Model{Map: map[string]interface{}{"resource": "orchestration", "message": clientErr.Error()}}
		_ = websupport.ModelAndView(w, &resources, "orchestration_new", model)
		log.Println(clientErr.Error())
		return
	}
	http.Redirect(w, r, "/orchestration/new", http.StatusMovedPermanently)
}
