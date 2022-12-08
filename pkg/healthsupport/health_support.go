package healthsupport

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"strconv"
)

type HealthCheck interface {
	Name() string
	Check() bool
}

type NoopCheck struct {
}

func (d *NoopCheck) Name() string {
	return "noop"
}

func (d *NoopCheck) Check() bool {
	return true
}

type response struct {
	Name string `json:"name"`
	Pass string `json:"pass"`
}

type HealthInfo struct {
	Status string `json:"status"`
}

type httpClient interface {
	Get(url string) (*http.Response, error)
}

func HealthHandlerFunction(w http.ResponseWriter, _ *http.Request) {
	data, _ := json.Marshal(&HealthInfo{"pass"})
	w.Header().Set("content-type", "application/json")
	w.WriteHeader(http.StatusOK)
	_, _ = w.Write(data)
}

func HealthHandlerFunctionWithChecks(w http.ResponseWriter, _ *http.Request, checks []HealthCheck) { // todo - move to func?
	responses := make([]response, 0)
	for _, check := range checks {
		responses = append(responses, response{
			Name: check.Name(),
			Pass: strconv.FormatBool(check.Check()),
		})
	}
	data, _ := json.Marshal(responses)
	w.Header().Set("content-type", "application/json")
	w.WriteHeader(http.StatusOK)
	_, _ = w.Write(data)
}

func WaitForHealthy(server *http.Server) {
	WaitForHealthyWithClient(server, http.DefaultClient, fmt.Sprintf("http://%s/health", server.Addr))
}

func WaitForHealthyWithClient(server *http.Server, client httpClient, url string) {
	var isLive bool
	for !isLive {
		resp, err := client.Get(url)
		if err == nil && resp.StatusCode == http.StatusOK {
			log.Println("Server is healthy.", server.Addr)
			isLive = true
		}
	}
}
