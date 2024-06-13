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

	"github.com/hexa-org/policy-mapper/pkg/hexapolicy"
	"github.com/hexa-org/policy-orchestrator/demo/pkg/oauth2support"
)

type HTTPClient interface {
	Get(url string) (resp *http.Response, err error)
	Do(req *http.Request) (*http.Response, error)
}

type orchestratorClient struct {
	client     HTTPClient
	jwtHandler oauth2support.JwtClientHandler
	url        string
}

// GetHttpClient used mainly for testing
func (c orchestratorClient) GetHttpClient() HTTPClient {
	return c.client
}

// NewOrchestratorClient returns a handle to a client that calls the Orchestrator service.
// When `client` is specified it will override OAuth2 token support.
func NewOrchestratorClient(client HTTPClient, url string) Client {
	var jwtHandler oauth2support.JwtClientHandler
	if client == nil {
		jwtHandler = oauth2support.NewJwtClientHandler()
		client = jwtHandler.GetHttpClient()
	}

	return &orchestratorClient{client, jwtHandler, url}
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

func (c orchestratorClient) Applications(refresh bool) (applications []Application, err error) {
	refreshParam := ""
	if refresh {
		refreshParam = "?refresh=true"
	}
	url := fmt.Sprintf("%v/applications%s", c.url, refreshParam)
	resp, hawkErr := c.client.Get(url)
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
	resp, hawkErr := c.client.Get(url)
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
	resp, hawkErr := c.client.Get(url)
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
	req, _ := http.NewRequest(http.MethodPost, url, bytes.NewBuffer(marshal))
	resp, hawkErr := c.client.Do(req)
	return errorOrBadResponse(resp, http.StatusCreated, hawkErr)
}

func (c orchestratorClient) DeleteIntegration(id string) error {
	url := fmt.Sprintf("%v/integrations/%s", c.url, id)
	resp, hawkErr := c.client.Get(url)
	return errorOrBadResponse(resp, http.StatusOK, hawkErr)
}

func (c orchestratorClient) GetPolicies(id string) ([]hexapolicy.PolicyInfo, string, error) {
	url := fmt.Sprintf("%v/applications/%s/policies", c.url, id)
	resp, hawkErr := c.client.Get(url)
	if err := errorOrBadResponse(resp, http.StatusOK, hawkErr); err != nil {
		return []hexapolicy.PolicyInfo{}, "{}", err
	}

	jsonBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return []hexapolicy.PolicyInfo{}, "{}", err
	}

	var jsonResponse hexapolicy.Policies
	if err := json.NewDecoder(bytes.NewReader(jsonBody)).Decode(&jsonResponse); err != nil {
		log.Println(err)
		return []hexapolicy.PolicyInfo{}, string(jsonBody), err
	}

	return jsonResponse.Policies, string(jsonBody), nil
}

func (c orchestratorClient) SetPolicies(id string, policies string) error {
	url := fmt.Sprintf("%v/applications/%s/policies", c.url, id)
	req, _ := http.NewRequest(http.MethodPost, url, strings.NewReader(policies))
	resp, err := c.client.Do(req)
	return errorOrBadResponse(resp, http.StatusCreated, err)
}

type orchestration struct {
	From string `json:"from"`
	To   string `json:"to"`
}

func (c orchestratorClient) Orchestration(from string, to string) error {
	url := fmt.Sprintf("%v/orchestration", c.url)
	marshal, _ := json.Marshal(orchestration{From: from, To: to})
	req, _ := http.NewRequest(http.MethodPost, url, bytes.NewReader(marshal))
	resp, err := c.client.Do(req)
	return errorOrBadResponse(resp, http.StatusCreated, err)
}

func errorOrBadResponse(response *http.Response, status int, err error) error {
	if err != nil {
		log.Println(err)
		return err
	}
	if response.StatusCode != status {
		err := checkTokenError(response)
		if err != nil {
			log.Println(err.Error())
			return err
		}
		all, _ := io.ReadAll(response.Body)
		message := string(all)
		log.Println(message)
		return errors.New(message)
	}
	return err
}

func checkTokenError(response *http.Response) error {
	if response.StatusCode == http.StatusUnauthorized || response.StatusCode == http.StatusForbidden {
		msg := response.Header.Get("WWW-Authenticate")
		if msg == "" {
			return errors.New(http.StatusText(response.StatusCode))
		}
		return errors.New(msg)
	}
	return nil
}
