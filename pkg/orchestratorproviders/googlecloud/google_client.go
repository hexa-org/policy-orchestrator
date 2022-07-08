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
	"strings"
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

type engines struct {
	ID              string `json:"id"`
	Name            string `json:"name"`
	DefaultHostname string `json:"defaultHostname"`
}

func (c *GoogleClient) GetAppEngineApplications() ([]orchestrator.ApplicationInfo, error) {
	url := fmt.Sprintf("https://appengine.googleapis.com/v1/apps/%s", c.ProjectId)
	var appEngines engines

	get, err := c.HttpClient.Get(url)
	if err != nil {
		log.Println("Unable to find google cloud app engine applications.")
		return []orchestrator.ApplicationInfo{}, err
	}
	log.Printf("Google cloud response %s.\n", get.Status)

	if err = json.NewDecoder(get.Body).Decode(&appEngines); err != nil {
		log.Println("Unable to decode google cloud app engine applications.")
		return []orchestrator.ApplicationInfo{}, err
	}

	log.Printf("Found google cloud backend app engine applications %s.\n", appEngines.Name)

	apps := []orchestrator.ApplicationInfo{
		{ObjectID: appEngines.ID, Name: appEngines.Name, Description: appEngines.DefaultHostname},
	}
	return apps, nil
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

func (c *GoogleClient) GetBackendPolicy(name, objectId string) ([]policysupport.PolicyInfo, error) {
	var url string
	if strings.HasPrefix(name, "k8s") { // todo - revisit and improve the decision here
		url = fmt.Sprintf("https://iap.googleapis.com/v1/projects/%s/iap_web/compute/services/%s:getIamPolicy", c.ProjectId, objectId)
	} else {
		url = fmt.Sprintf("https://iap.googleapis.com/v1/projects/%s/iap_web/appengine-%s/services/default:getIamPolicy", c.ProjectId, objectId)
	}

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
			Meta:    policysupport.MetaInfo{Version: "0.5"},
			Actions: []policysupport.ActionInfo{{"gcp:" + found.Role}},
			Subject: policysupport.SubjectInfo{Members: found.Members},
			Object: policysupport.ObjectInfo{
				ResourceID: objectId,
			},
		})
	}
	return policies, err
}

func (c *GoogleClient) SetBackendPolicy(name, objectId string, p policysupport.PolicyInfo) error { // todo - objectId may no longer be needed, at least for google
	var url string
	if strings.HasPrefix(name, "k8s") { // todo - revisit and improve the decision here
		url = fmt.Sprintf("https://iap.googleapis.com/v1/projects/%s/iap_web/compute/services/%s:setIamPolicy", c.ProjectId, objectId)
	} else {
		url = fmt.Sprintf("https://iap.googleapis.com/v1/projects/%s/iap_web/appengine-%s/services/default:setIamPolicy", c.ProjectId, objectId)
	}

	// todo - handle many actions
	uri := strings.TrimPrefix(p.Actions[0].ActionUri, "gcp:")

	body := policy{bindings{[]bindingInfo{{uri, p.Subject.Members}}}}
	b := new(bytes.Buffer)
	_ = json.NewEncoder(b).Encode(body)

	_, err := c.HttpClient.Post(url, "application/json", b)
	return err
}
