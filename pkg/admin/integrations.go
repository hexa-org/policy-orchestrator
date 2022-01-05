package admin

import (
	"hexa/pkg/web_support"
	"log"
	"net/http"
)

type integrationsHandler struct{}

func NewIntegrationsHandler() integrationsHandler { return integrationsHandler{} }

func (i integrationsHandler) IntegrationsHandler(w http.ResponseWriter, r *http.Request) {
	model := web_support.Model{Map: map[string]interface{}{}}
	err := web_support.ModelAndView(w, "discovery", model)
	if err != nil {
		log.Println(err)
		return
	}
}
