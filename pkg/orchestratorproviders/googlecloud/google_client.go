package googlecloud

import (
	"bytes"
	"encoding/json"
	"fmt"
	"github.com/hexa-org/policy-orchestrator/pkg/orchestrator"
	"github.com/hexa-org/policy-orchestrator/pkg/policysupport"
	"io"
	"log"
	"net/http"
)

type HTTPClient interface {
	Get(url string) (resp *http.Response, err error)
	Post(url, contentType string, body io.Reader) (resp *http.Response, err error)
}

type GoogleClient struct {
	HttpClient HTTPClient
	ProjectId  string
}

type backends struct {
	ID        string        `json:"id"`
	Resources []backendInfo `json:"items"`
}

type backendInfo struct {
	ID          string `json:"id"`
	Name        string `json:"name"`
	Description string `json:"description"`
}

func (c *GoogleClient) GetBackendApplications() ([]orchestrator.ApplicationInfo, error) {
	url := fmt.Sprintf("https://compute.googleapis.com/compute/v1/projects/%s/global/backendServices", c.ProjectId)

	get, err := c.HttpClient.Get(url)
	if err != nil {
		log.Println("Unable to find google cloud backend services.")
		return []orchestrator.ApplicationInfo{}, err
	}
	log.Printf("Google cloud response %s.\n", get.Status)

	var backend backends
	if err = json.NewDecoder(get.Body).Decode(&backend); err != nil {
		log.Println("Unable to decode google cloud backend services.")
		return []orchestrator.ApplicationInfo{}, err
	}

	var apps []orchestrator.ApplicationInfo
	for _, info := range backend.Resources {
		log.Printf("Found google cloud backend services %s.\n", info.Name)
		apps = append(apps, orchestrator.ApplicationInfo{ObjectID: info.ID, Name: info.Name, Description: info.Description})
	}
	return apps, nil
}

type policy struct {
	Policy bindings `json:"policy"`
}

type bindings struct {
	Bindings []bindingInfo `json:"bindings"`
}

type bindingInfo struct {
	Role    string   `json:"role"`
	Members []string `json:"members"`
}

func (c *GoogleClient) GetBackendPolicy(objectId string) ([]policysupport.PolicyInfo, error) {
	url := fmt.Sprintf("https://iap.googleapis.com/v1/projects/%s/iap_web/compute/services/%s:getIamPolicy", c.ProjectId, objectId)

	post, err := c.HttpClient.Post(url, "application/json", bytes.NewReader([]byte{}))
	if err != nil {
		log.Println("Unable to find google cloud policy.")
		return []policysupport.PolicyInfo{}, err
	}
	log.Printf("Google cloud response %s.\n", post.Status)

	var binds bindings
	if err = json.NewDecoder(post.Body).Decode(&binds); err != nil {
		log.Println("Unable to decode google cloud policy.")
		return []policysupport.PolicyInfo{}, err
	}

	var policies []policysupport.PolicyInfo
	for _, found := range binds.Bindings {
		log.Printf("Found google cloud policy for role %s.\n", found.Role)
		policies = append(policies, policysupport.PolicyInfo{
			Version: "0.1",
			Action:  found.Role,
			Subject: policysupport.SubjectInfo{AuthenticatedUsers: found.Members},
			Object:  policysupport.ObjectInfo{Resources: []string{"/"}},
		})
	}
	return policies, err
}

func (c *GoogleClient) SetBackendPolicy(objectId string, p policysupport.PolicyInfo) error {
	url := fmt.Sprintf("https://iap.googleapis.com/v1/projects/%s/iap_web/compute/services/%s:setIamPolicy", c.ProjectId, objectId)

	body := policy{bindings{[]bindingInfo{{p.Action, p.Subject.AuthenticatedUsers}}}}
	b := new(bytes.Buffer)
	_ = json.NewEncoder(b).Encode(body)

	_, err := c.HttpClient.Post(url, "application/json", b)
	return err
}
