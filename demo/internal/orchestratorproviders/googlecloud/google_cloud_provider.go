package googlecloud

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"strings"

	"github.com/go-playground/validator/v10"
	"github.com/hexa-org/policy-mapper/pkg/hexapolicy"
	"github.com/hexa-org/policy-orchestrator/demo/internal/orchestrator"
	"google.golang.org/api/option"
	"google.golang.org/api/transport/http"
)

type GoogleProvider struct {
	HttpClientOverride HTTPClient
}

func (g *GoogleProvider) Name() string {
	return "google_cloud"
}

func (g *GoogleProvider) Project(key []byte) string {
	return g.credentials(key).ProjectId
}

func (g *GoogleProvider) DiscoverApplications(info orchestrator.IntegrationInfo) (apps []orchestrator.ApplicationInfo, err error) {
	if !strings.EqualFold(info.Name, g.Name()) {
		return apps, err
	}

	key := info.Key
	foundCredentials := g.credentials(key)
	client, createClientErr := g.getHttpClient(key)
	if createClientErr != nil {
		fmt.Println("Unable to create google http client.")
		return apps, createClientErr
	}
	googleClient := GoogleClient{client, foundCredentials.ProjectId}

	backendApplications, _ := googleClient.GetBackendApplications()
	apps = append(apps, backendApplications...)

	appEngineApplications, _ := googleClient.GetAppEngineApplications()
	apps = append(apps, appEngineApplications...)

	return apps, err
}

func (g *GoogleProvider) GetPolicyInfo(integration orchestrator.IntegrationInfo, app orchestrator.ApplicationInfo) (infos []hexapolicy.PolicyInfo, err error) {
	key := integration.Key
	foundCredentials := g.credentials(key)
	client, createClientErr := g.getHttpClient(key)
	if createClientErr != nil {
		fmt.Println("Unable to create google http client.")
		return infos, createClientErr
	}
	googleClient := GoogleClient{client, foundCredentials.ProjectId}
	return googleClient.GetBackendPolicy(app.Name, app.ObjectID)
}

func (g *GoogleProvider) SetPolicyInfo(integration orchestrator.IntegrationInfo, app orchestrator.ApplicationInfo, policyInfos []hexapolicy.PolicyInfo) (int, error) {
	validate := validator.New() // todo - move this up?
	errApp := validate.Struct(app)
	if errApp != nil {
		return 500, errApp
	}
	errPolicies := validate.Var(policyInfos, "omitempty,dive")
	if errPolicies != nil {
		return 500, errPolicies
	}

	key := integration.Key
	foundCredentials := g.credentials(key)
	client, createClientErr := g.getHttpClient(key)
	if createClientErr != nil {
		fmt.Println("Unable to create google http client.")
		return 500, createClientErr
	}
	googleClient := GoogleClient{client, foundCredentials.ProjectId}
	for _, policyInfo := range policyInfos {
		err := googleClient.SetBackendPolicy(app.Name, app.ObjectID, policyInfo)
		if err != nil {
			return 500, err
		}
	}
	return 201, nil
}

func (g *GoogleProvider) NewHttpClient(key []byte) (HTTPClient, error) {
	var opts []option.ClientOption
	opt := option.WithCredentialsJSON(key)
	opts = append([]option.ClientOption{option.WithScopes("https://www.googleapis.com/auth/cloud-platform")}, opt)
	client, _, err := http.NewClient(context.Background(), opts...)
	return client, err
}

type credentials struct {
	ProjectId string `json:"project_id"`
}

func (g *GoogleProvider) credentials(key []byte) credentials {
	var foundCredentials credentials
	_ = json.NewDecoder(bytes.NewReader(key)).Decode(&foundCredentials)
	return foundCredentials
}

func (g *GoogleProvider) getHttpClient(key []byte) (HTTPClient, error) {
	if g.HttpClientOverride != nil {
		return g.HttpClientOverride, nil
	}
	return g.NewHttpClient(key)
}
