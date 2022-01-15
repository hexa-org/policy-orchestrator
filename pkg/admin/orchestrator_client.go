package admin

import (
	"bytes"
	"encoding/json"
	"github.com/hexa-org/policy-orchestrator/pkg/hawksupport"
	"io"
	"log"
	"net/http"
)

type HTTPClient interface {
	Get(url string) (resp *http.Response, err error)
	Do(req *http.Request) (*http.Response, error)
}

type orchestratorClient struct {
	client HTTPClient
	key    string
}

func NewOrchestratorClient(client HTTPClient, key string) Client {
	return &orchestratorClient{client, key}
}

func (c orchestratorClient) Health(url string) (string, error) {
	resp, err := c.client.Get(url)
	if err != nil {
		log.Println(err)
		return "{\"status\": \"fail\"}", err
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
}

func (c orchestratorClient) Applications(url string) (applications []Application, err error) {
	resp, err := hawksupport.HawkGet(c.client, "anId", c.key, url)
	if err != nil {
		log.Println(err)
		return applications, err
	}

	var jsonResponse applicationList
	body := resp.Body
	err = json.NewDecoder(body).Decode(&jsonResponse)
	if err != nil {
		log.Printf("unable to parse customer json: %s\n", err.Error())
		return applications, err
	}

	for _, app := range jsonResponse.Applications {
		applications = append(applications, Application{
			app.ID,
			app.IntegrationId,
			app.ObjectId,
			app.Name,
			app.Description})
	}

	return applications, nil
}

func (c orchestratorClient) Application(url string) (Application, error) {
	resp, err := hawksupport.HawkGet(c.client, "anId", c.key, url)
	if err != nil {
		return Application{}, err
	}

	var jsonResponse application
	body := resp.Body
	err = json.NewDecoder(body).Decode(&jsonResponse)
	if err != nil {
		return Application{}, err
	}
	app := Application{jsonResponse.ID, jsonResponse.IntegrationId, jsonResponse.ObjectId, jsonResponse.Name, jsonResponse.Description}
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

func (c orchestratorClient) Integrations(url string) (integrations []Integration, err error) {
	resp, err := hawksupport.HawkGet(c.client, "anId", c.key, url)
	if err != nil {
		return integrations, err
	}

	var jsonResponse integrationList
	body := resp.Body
	err = json.NewDecoder(body).Decode(&jsonResponse)
	if err != nil {
		return integrations, err
	}

	for _, in := range jsonResponse.Integrations {
		integrations = append(integrations, Integration{in.ID, in.Name, in.Provider, []byte(in.Key)})
	}

	return integrations, nil
}

func (c orchestratorClient) CreateIntegration(url string, provider string, key []byte) error {
	integration := integration{Name: "aName", Provider: provider, Key: key} // todo - replace with project name in key
	marshal, _ := json.Marshal(integration)
	_, err := hawksupport.HawkPost(c.client, "anId", c.key, url, bytes.NewReader(marshal))
	return err
}

func (c orchestratorClient) DeleteIntegration(url string) error {
	_, err := hawksupport.HawkGet(c.client, "anId", c.key, url)
	return err
}
