package admin

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"strings"

	"github.com/hexa-org/policy-orchestrator/pkg/websupport"
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
	health, _ := p.client.Health()

	var checks []Check
	if err := json.NewDecoder(strings.NewReader(health)).Decode(&checks); err != nil {
		log.Printf("unable to parse found json for status check: %s\n", err.Error())
		checks = append(checks, Check{"Unparsable", "false"})
	}
	status := Status{fmt.Sprintf("%v/health", p.orchestratorUrl), checks} // todo - remove endpoint knowledge

	model := websupport.Model{Map: map[string]interface{}{"resource": "status", "status": status}}
	_ = websupport.ModelAndView(w, &resources, "status", model)
}
