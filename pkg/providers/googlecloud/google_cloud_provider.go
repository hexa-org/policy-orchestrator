package googlecloud

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"github.com/hexa-org/policy-orchestrator/pkg/identityquerylanguage"
	"github.com/hexa-org/policy-orchestrator/pkg/orchestrator/provider"
	"google.golang.org/api/option"
	"google.golang.org/api/transport/http"
	"strings"
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

func (g *GoogleProvider) DiscoverApplications(info provider.IntegrationInfo) (apps []provider.ApplicationInfo, err error) {
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
	found, _ := googleClient.GetBackendApplications()
	apps = append(apps, found...)
	return apps, err
}

func (g *GoogleProvider) GetPolicyInfo(integration provider.IntegrationInfo, app provider.ApplicationInfo) (infos []identityquerylanguage.PolicyInfo, err error) {
	key := integration.Key
	foundCredentials := g.credentials(key)
	client, createClientErr := g.getHttpClient(key)
	if createClientErr != nil {
		fmt.Println("Unable to create google http client.")
		return infos, createClientErr
	}
	googleClient := GoogleClient{client, foundCredentials.ProjectId}
	return googleClient.GetBackendPolicy(app.ObjectID)
}

func (g *GoogleProvider) SetPolicyInfo(integration provider.IntegrationInfo, app provider.ApplicationInfo, policies []identityquerylanguage.PolicyInfo) error {
	key := integration.Key
	foundCredentials := g.credentials(key)
	client, createClientErr := g.getHttpClient(key)
	if createClientErr != nil {
		fmt.Println("Unable to create google http client.")
		return createClientErr
	}
	googleClient := GoogleClient{client, foundCredentials.ProjectId}
	for _, policyInfo := range policies {
		err := googleClient.SetBackendPolicy(app.ObjectID, policyInfo)
		if err != nil {
			return err
		}
	}
	return nil
}

func (g *GoogleProvider) NewHttpClient(key []byte) (HTTPClient, error) {
	var opts []option.ClientOption
	opt := option.WithCredentialsJSON(key)
	opts = append([]option.ClientOption{option.WithScopes("https://www.googleapis.com/auth/cloud-platform")}, opt)
	client, _, err := http.NewClient(context.Background(), opts...)
	return client, err
}

///

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
