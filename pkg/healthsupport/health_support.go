package healthsupport

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
)

type HealthInfo struct {
	Status string `json:"status"`
}

func HealthHandlerFunction(w http.ResponseWriter, _ *http.Request) {
	data, _ := json.Marshal(&HealthInfo{"pass"})
	w.Header().Set("content-type", "application/json")
	w.WriteHeader(http.StatusOK)
	_, _ = w.Write(data)
}

func WaitForHealthy(server *http.Server) {
	WaitForHealthyWithClient(server, &http.Client{}, fmt.Sprintf("http://%s/health", server.Addr))
}

func WaitForHealthyWithClient(server *http.Server, client *http.Client, url string) {
	var isLive bool
	for !isLive {
		resp, err := client.Get(url)
		if err == nil && resp.StatusCode == http.StatusOK {
			log.Println("Server is healthy.", server.Addr)
			isLive = true
		}
	}
}
