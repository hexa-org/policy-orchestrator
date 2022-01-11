package google_cloud

import (
	"encoding/json"
	"fmt"
	"hexa/pkg/orchestrator/provider"
	"log"
	"net/http"
)

type HTTPClient interface {
	Get(url string) (resp *http.Response, err error)
}

type backendInfo struct {
	ID          string `json:"id"`
	Name        string `json:"name"`
	Description string `json:"description"`
}

type backends struct {
	ID        string        `json:"id"`
	Resources []backendInfo `json:"items"`
}

type GoogleClient struct {
	HttpClient HTTPClient
	ProjectId  string
}

func (c *GoogleClient) GetBackendApplications() (apps []provider.ApplicationInfo, err error) {
	get, err := c.HttpClient.Get(fmt.Sprintf("https://compute.googleapis.com/compute/v1/projects/%s/global/backendServices", c.ProjectId))
	if err != nil {
		log.Println("Unable to find google cloud backend services.")
		return apps, err
	}
	log.Printf("Google cloud response %s.\n", get.Status)
	var backend backends
	err = json.NewDecoder(get.Body).Decode(&backend)
	if err != nil {
		log.Println("Unable to decode google cloud backend services.")
		return apps, err
	}
	for _, info := range backend.Resources {
		log.Printf("Found google cloud backend services %s.\n", info.Name)
		apps = append(apps, provider.ApplicationInfo{ID: info.ID, Name: info.Name, Description: info.Description})
	}
	return apps, nil
}
