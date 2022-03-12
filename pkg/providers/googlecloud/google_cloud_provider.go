package googlecloud

import (
	"bytes"
	"context"
	"encoding/json"
	"github.com/hexa-org/policy-orchestrator/pkg/orchestrator/provider"
	"google.golang.org/api/option"
	"google.golang.org/api/transport/http"
	"strings"
)

type GoogleProvider struct {
	Http HTTPClient
}

func (g *GoogleProvider) Name() string {
	return "google_cloud"
}

func (g *GoogleProvider) Project(key []byte) string {
	return g.credentials(key).ProjectId
}

func (g *GoogleProvider) DiscoverApplications(info provider.IntegrationInfo) (apps []provider.ApplicationInfo, err error) {
	key := info.Key
	foundCredentials := g.credentials(key)
	if strings.EqualFold(info.Name, g.Name()) {
		g.ensureClientIsAvailable(key)
		googleClient := GoogleClient{g.Http, foundCredentials.ProjectId}
		found, _ := googleClient.GetBackendApplications()
		apps = append(apps, found...)
	}
	return apps, err
}

func (g *GoogleProvider) GetPolicyInfo(integration provider.IntegrationInfo, app provider.ApplicationInfo) (infos []provider.PolicyInfo, err error) {
	key := integration.Key
	foundCredentials := g.credentials(key)
	g.ensureClientIsAvailable(key)
	googleClient := GoogleClient{g.Http, foundCredentials.ProjectId}
	return googleClient.GetBackendPolicy(app.ObjectID)
}

func (g *GoogleProvider) SetPolicyInfo(integration provider.IntegrationInfo, app provider.ApplicationInfo, policies []provider.PolicyInfo) error {
	key := integration.Key
	foundCredentials := g.credentials(key)
	g.ensureClientIsAvailable(key)
	googleClient := GoogleClient{g.Http, foundCredentials.ProjectId}
	for _, policyInfo := range policies {
		err := googleClient.SetBackendPolicy(app.ObjectID, policyInfo)
		if err != nil {
			return err
		}
	}
	return nil
}

func (g *GoogleProvider) HttpClient(key []byte) (HTTPClient, error) {
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

func (g *GoogleProvider) ensureClientIsAvailable(key []byte) {
	if g.Http == nil {
		g.Http, _ = g.HttpClient(key) // todo - for testing, might be a better way?
	}
}