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
	var isLive bool
	for !isLive {
		resp, err := http.Get(fmt.Sprintf("http://%s/health", server.Addr))
		if err == nil && resp.StatusCode == http.StatusOK {
			log.Println("Server is healthy.", server.Addr)
			isLive = true
		}
	}
}
