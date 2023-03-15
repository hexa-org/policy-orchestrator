package googlecloud

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"strings"

	"github.com/hexa-org/policy-orchestrator/internal/orchestrator"
	"github.com/hexa-org/policy-orchestrator/internal/policysupport"
	"github.com/hexa-org/policy-orchestrator/pkg/functionalsupport"
	"github.com/hexa-org/policy-orchestrator/pkg/googlesupport"
	"github.com/hexa-org/policy-orchestrator/pkg/hexapolicy"
	"google.golang.org/api/iam/v1"
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

	if get.StatusCode == 404 {
		log.Println("No App Engine Found")
		return []orchestrator.ApplicationInfo{}, nil
	}

	log.Printf("Google cloud response %s.\n", get.Status)

	if err = json.NewDecoder(get.Body).Decode(&appEngines); err != nil {
		log.Println("Unable to decode google cloud app engine applications.")
		return []orchestrator.ApplicationInfo{}, err
	}

	log.Printf("Found google cloud backend app engine applications %s.\n", appEngines.Name)

	apps := []orchestrator.ApplicationInfo{
		{ObjectID: appEngines.ID, Name: appEngines.Name, Description: appEngines.DefaultHostname, Service: "AppEngine"},
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
		var service string
		if strings.HasPrefix(info.Name, "k8s") {
			service = "Kubernetes"
		} else {
			service = "Cloud Run"
		}
		apps = append(apps, orchestrator.ApplicationInfo{ObjectID: info.ID, Name: info.Name, Description: info.Description, Service: service})
	}
	return apps, nil
}

type policy struct {
	Policy bindings `json:"policy"`
}

type bindings struct {
	Bindings []bindingInfo `json:"bindings"`
}

type condition struct {
	Title       string `json:"title"`
	Description string `json:"description"`
	Expression  string `json:"expression"`
}

type bindingInfo struct {
	Role      string     `json:"role"`
	Members   []string   `json:"members"`
	Condition *condition `json:"condition,omitempty"`
}

func (c *GoogleClient) GetBackendPolicy(name, objectId string) ([]policysupport.PolicyInfo, error) {
	var url string
	if strings.HasPrefix(name, "apps") { // todo - revisit and improve the decision here
		url = fmt.Sprintf("https://iap.googleapis.com/v1/projects/%s/iap_web/appengine-%s/services/default:getIamPolicy", c.ProjectId, objectId)
	} else {
		url = fmt.Sprintf("https://iap.googleapis.com/v1/projects/%s/iap_web/compute/services/%s:getIamPolicy", c.ProjectId, objectId)
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

	/// todo - below is work in progress

	iamBindings := functionalsupport.Map(binds.Bindings, func(binding bindingInfo) iam.Binding {
		return iam.Binding{
			Condition:       nil,
			Members:         binding.Members,
			Role:            binding.Role,
			ForceSendFields: nil,
			NullFields:      nil,
		}
	})

	policies := functionalsupport.Map(iamBindings, func(iamBinding iam.Binding) hexapolicy.PolicyInfo {
		p, mappingErr := googlesupport.New(map[string]string{}).MapBindingToPolicy(objectId, iamBinding)
		if mappingErr != nil {
			return hexapolicy.PolicyInfo{}
		}
		return p
	})

	// todo - use mapper policy support here...
	hexaPolicies := functionalsupport.Map(policies, func(policy hexapolicy.PolicyInfo) policysupport.PolicyInfo {
		return policysupport.PolicyInfo{
			Meta: policysupport.MetaInfo{Version: policy.Meta.Version},
			Actions: functionalsupport.Map(policy.Actions, func(action hexapolicy.ActionInfo) policysupport.ActionInfo {
				return policysupport.ActionInfo{
					ActionUri: action.ActionUri,
				}
			}),
			Subject: policysupport.SubjectInfo{Members: policy.Subject.Members},
			Object:  policysupport.ObjectInfo{ResourceID: policy.Object.ResourceID},
		}
	})
	return hexaPolicies, err
}

func (c *GoogleClient) SetBackendPolicy(name, objectId string, p policysupport.PolicyInfo) error { // todo - objectId may no longer be needed, at least for google
	var url string
	if strings.HasPrefix(name, "apps") { // todo - revisit and improve the decision here
		url = fmt.Sprintf("https://iap.googleapis.com/v1/projects/%s/iap_web/appengine-%s/services/default:setIamPolicy", c.ProjectId, objectId)
	} else {
		url = fmt.Sprintf("https://iap.googleapis.com/v1/projects/%s/iap_web/compute/services/%s:setIamPolicy", c.ProjectId, objectId)
	}

	// todo - handle many actions
	uri := strings.TrimPrefix(p.Actions[0].ActionUri, "gcp:")

	body := policy{Policy: bindings{[]bindingInfo{{Role: uri, Members: p.Subject.Members}}}}
	b := new(bytes.Buffer)
	_ = json.NewEncoder(b).Encode(body)

	_, err := c.HttpClient.Post(url, "application/json", b)
	return err
}
