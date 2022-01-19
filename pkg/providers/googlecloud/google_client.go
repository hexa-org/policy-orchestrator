package googlecloud

import (
	"bytes"
	"encoding/json"
	"fmt"
	"github.com/hexa-org/policy-orchestrator/pkg/orchestrator/provider"
	"io"
	"log"
	"net/http"
)

type HTTPClient interface {
	Get(url string) (resp *http.Response, err error)
	Post(url, contentType string, body io.Reader) (resp *http.Response, err error)
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

type bindings struct {
	Bindings []bindingInfo `json:"bindings"`
}

type bindingInfo struct {
	Role    string   `json:"role"`
	Members []string `json:"members"`
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

func (c *GoogleClient) GetBackendPolicy(objectId string) (infos []provider.PolicyInfo, err error) {
	var b []byte
	url := fmt.Sprintf("https://iap.googleapis.com/v1/projects/%s/iap_web/compute/services/%s", c.ProjectId, objectId)
	post, err := c.HttpClient.Post(url, "application/json", bytes.NewReader(b))
	if err != nil {
		log.Println("Unable to find google cloud policy.")
		return infos, err
	}
	log.Printf("Google cloud response %s.\n", post.Status)

	var binds bindings
	err = json.NewDecoder(post.Body).Decode(&binds)
	if err != nil {
		log.Println("Unable to decode google cloud policy.")
		return infos, err
	}

	for _, found := range binds.Bindings {
		log.Printf("Found google cloud policy for role %s.\n", found.Role)
		infos = append(infos, provider.PolicyInfo{
			Version: "0.1",
			Action:  found.Role,
			Subject: provider.SubjectInfo{AuthenticatedUsers: found.Members},
			Object:  provider.ObjectInfo{Resources: []string{"/"}},
		})
	}
	return infos, err
}

func (c *GoogleClient) SetBackendPolicy() {
	//https: //iap.googleapis.com/v1/{resource=**}:setIamPolicy
}
