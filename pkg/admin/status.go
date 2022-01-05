package admin

import (
	"fmt"
	"hexa/pkg/web_support"
	"log"
	"net/http"
)

type Status struct {
	URL    string
	Status string
}

type statusHandler struct {
	orchestratorUrl string
	client           Client
}

func NewStatusHandler(orchestratorUrl string, client Client) statusHandler {
	return statusHandler{orchestratorUrl, client}
}

func (p statusHandler) StatusHandler(w http.ResponseWriter, r *http.Request) {
	url := fmt.Sprintf("%v/health", p.orchestratorUrl)
	health, _ := p.client.Health(url)
	status := Status{url, health}
	model := web_support.Model{Map: map[string]interface{}{"status": status}}
	err := web_support.ModelAndView(w, "status", model)
	if err != nil {
		log.Println(err)
		return
	}
}
