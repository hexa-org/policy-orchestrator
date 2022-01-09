package admin

import (
	"fmt"
	"hexa/pkg/web_support"
	"net/http"
)

type Application struct {
	Name string
}

type applicationsHandler struct {
	orchestratorUrl string
	client          Client
}

func NewApplicationsHandler(orchestratorUrl string, client Client) applicationsHandler {
	return applicationsHandler{orchestratorUrl, client}
}

func (p applicationsHandler) ApplicationsHandler(w http.ResponseWriter, r *http.Request) {
	url := fmt.Sprintf("%v/applications", p.orchestratorUrl)
	applications, _ := p.client.Applications(url)
	model := web_support.Model{Map: map[string]interface{}{"resource": "applications", "applications": applications}}
	_ = web_support.ModelAndView(w, "applications", model)
}
