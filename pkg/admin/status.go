package admin

import (
	"encoding/json"
	"fmt"
	"github.com/hexa-org/policy-orchestrator/pkg/websupport"
	"log"
	"net/http"
	"strings"
)

type Status struct {
	URL    string
	Checks []Check
}

type StatusHandler struct {
	orchestratorUrl string
	client          Client
}

func NewStatusHandler(orchestratorUrl string, client Client) StatusHandler {
	return StatusHandler{orchestratorUrl, client}
}

type Check struct {
	Name string `json:"name"`
	Pass string `json:"pass"`
}

func (p StatusHandler) StatusHandler(w http.ResponseWriter, _ *http.Request) {
	url := fmt.Sprintf("%v/health", p.orchestratorUrl)
	health, _ := p.client.Health(url)

	var checks []Check
	if err := json.NewDecoder(strings.NewReader(health)).Decode(&checks); err != nil {
		log.Printf("unable to parse found json for status check: %s\n", err.Error())
		checks = append(checks, Check{"Unparsable", "false"})
	}
	status := Status{url, checks}

	model := websupport.Model{Map: map[string]interface{}{"resource": "status", "status": status}}
	_ = websupport.ModelAndView(w, "status", model)
}
