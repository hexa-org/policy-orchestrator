package admin

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log"
	"net/http"
	"strings"

	"github.com/hexa-org/policy-orchestrator/pkg/hawksupport"
)

type HTTPClient interface {
	Get(url string) (resp *http.Response, err error)
	Do(req *http.Request) (*http.Response, error)
}

type orchestratorClient struct {
	client HTTPClient
	url    string
	key    string
}

func NewOrchestratorClient(client HTTPClient, url, key string) Client {
	return &orchestratorClient{client, url, key}
}

func (c orchestratorClient) Health() (string, error) {
	resp, err := c.client.Get(fmt.Sprintf("%v/health", c.url))
	if err != nil {
		log.Println(err)
		return "[{\"name\":\"Unreachable\",\"pass\":\"fail\"}]", err
	}
	body, err := io.ReadAll(resp.Body)
	return string(body), err
}

type applicationList struct {
	Applications []application `json:"applications"`
}

type application struct {
	ID            string `json:"id"`
	IntegrationId string `json:"integration_id"`
	ObjectId      string `json:"object_id"`
	Name          string `json:"name"`
	Description   string `json:"description"`
	ProviderName  string `json:"provider_name"`
	Service       string `json:"service"`
}

func (c orchestratorClient) Applications() (applications []Application, err error) {
	url := fmt.Sprintf("%v/applications", c.url)
	resp, hawkErr := hawksupport.HawkGet(c.client, "anId", c.key, url)
	if err = errorOrBadResponse(resp, http.StatusOK, hawkErr); err != nil {
		return applications, err
	}

	var jsonResponse applicationList
	if err = json.NewDecoder(resp.Body).Decode(&jsonResponse); err != nil {
		log.Printf("unable to parse found json: %s\n", err.Error())
		return applications, err
	}

	for _, app := range jsonResponse.Applications {
		applications = append(applications, Application{
			ID:            app.ID,
			IntegrationId: app.IntegrationId,
			ObjectId:      app.ObjectId,
			Name:          app.Name,
			Description:   app.Description,
			ProviderName:  app.ProviderName,
			Service:       app.Service,
		})
	}

	return applications, nil
}

func (c orchestratorClient) Application(id string) (Application, error) {
	url := fmt.Sprintf("%v/applications/%s", c.url, id)
	resp, hawkErr := hawksupport.HawkGet(c.client, "anId", c.key, url)
	if err := errorOrBadResponse(resp, http.StatusOK, hawkErr); err != nil {
		return Application{}, err
	}

	var jsonResponse application
	if err := json.NewDecoder(resp.Body).Decode(&jsonResponse); err != nil {
		log.Printf("unable to parse found json: %s\n", err.Error())
		return Application{}, err
	}
	app := Application{
		ID:            jsonResponse.ID,
		IntegrationId: jsonResponse.IntegrationId,
		ObjectId:      jsonResponse.ObjectId,
		Name:          jsonResponse.Name,
		Description:   jsonResponse.Description,
		Service:       jsonResponse.Service,
	}
	return app, nil
}

type integrationList struct {
	Integrations []integration `json:"integrations"`
}

type integration struct {
	ID       string `json:"id"`
	Name     string `json:"name"`
	Provider string `json:"provider"`
	Key      []byte `json:"key"`
}

func (c orchestratorClient) Integrations() (integrations []Integration, err error) {
	url := fmt.Sprintf("%v/integrations", c.url)
	resp, hawkErr := hawksupport.HawkGet(c.client, "anId", c.key, url)
	if err = errorOrBadResponse(resp, http.StatusOK, hawkErr); err != nil {
		return integrations, err
	}

	var jsonResponse integrationList
	if err = json.NewDecoder(resp.Body).Decode(&jsonResponse); err != nil {
		log.Printf("unable to parse found json: %s\n", err.Error())
		return integrations, err
	}

	for _, in := range jsonResponse.Integrations {
		integrations = append(integrations, Integration{in.ID, in.Name, in.Provider, in.Key})
	}
	return integrations, nil
}

func (c orchestratorClient) CreateIntegration(name string, provider string, key []byte) error {
	url := fmt.Sprintf("%v/integrations", c.url)
	marshal, _ := json.Marshal(integration{Name: name, Provider: provider, Key: key})
	resp, hawkErr := hawksupport.HawkPost(c.client, "anId", c.key, url, bytes.NewReader(marshal))
	return errorOrBadResponse(resp, http.StatusCreated, hawkErr)
}

func (c orchestratorClient) DeleteIntegration(id string) error {
	url := fmt.Sprintf("%v/integrations/%s", c.url, id)
	resp, hawkErr := hawksupport.HawkGet(c.client, "anId", c.key, url)
	return errorOrBadResponse(resp, http.StatusOK, hawkErr)
}

type policies struct {
	Policies []policy `json:"policies"`
}

type policy struct {
	Meta    meta     `json:"meta"`
	Actions []action `json:"actions"`
	Subject subject  `json:"subject"`
	Object  object   `json:"object"`
}

type meta struct {
	Version string `json:"version"`
}

type action struct {
	ActionUri string `json:"action_uri"`
}

type subject struct {
	Members []string `json:"members"`
}

type object struct {
	ResourceId string   `json:"resource_id"`
	Resources  []string `json:"resources"`
}

func (c orchestratorClient) GetPolicies(id string) ([]Policy, string, error) {
	url := fmt.Sprintf("%v/applications/%s/policies", c.url, id)
	resp, hawkErr := hawksupport.HawkGet(c.client, "anId", c.key, url)
	if err := errorOrBadResponse(resp, http.StatusOK, hawkErr); err != nil {
		return []Policy{}, "{}", err
	}

	jsonBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return []Policy{}, "{}", err
	}

	var jsonResponse policies
	if err := json.NewDecoder(bytes.NewReader(jsonBody)).Decode(&jsonResponse); err != nil {
		log.Println(err)
		return []Policy{}, string(jsonBody), err
	}

	var foundPolicies []Policy
	for _, p := range jsonResponse.Policies {
		var actions []Action
		for _, a := range p.Actions {
			actions = append(actions, Action{a.ActionUri})
		}
		foundPolicies = append(foundPolicies, Policy{
			Meta:    Meta{p.Meta.Version},
			Actions: actions,
			Subject: Subject{Members: p.Subject.Members},
			Object:  Object{ResourceID: p.Object.ResourceId},
		})
	}
	return foundPolicies, string(jsonBody), nil
}

func (c orchestratorClient) SetPolicies(id string, policies string) error {
	url := fmt.Sprintf("%v/applications/%s/policies", c.url, id)
	resp, err := hawksupport.HawkPost(c.client, "anId", c.key, url, strings.NewReader(policies))
	return errorOrBadResponse(resp, http.StatusCreated, err)
}

type orchestration struct {
	From string `json:"from"`
	To   string `json:"to"`
}

func (c orchestratorClient) Orchestration(from string, to string) error {
	url := fmt.Sprintf("%v/orchestration", c.url)
	marshal, _ := json.Marshal(orchestration{From: from, To: to})
	resp, err := hawksupport.HawkPost(c.client, "anId", c.key, url, bytes.NewReader(marshal))
	return errorOrBadResponse(resp, http.StatusCreated, err)
}

func errorOrBadResponse(response *http.Response, status int, err error) error {
	if err != nil {
		log.Println(err)
		return err
	}
	if response.StatusCode != status {
		all, _ := io.ReadAll(response.Body)
		message := string(all)
		log.Println(message)
		return errors.New(message)
	}
	return err
}
