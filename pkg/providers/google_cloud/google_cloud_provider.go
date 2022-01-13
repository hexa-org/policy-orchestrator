package google_cloud

import (
	"bytes"
	"context"
	"encoding/json"
	"github.com/hexa-org/policy-orchestrator/pkg/orchestrator/provider"
	"google.golang.org/api/option"
	"google.golang.org/api/transport/http"
	"strings"
)

type credentials struct {
	ProjectId string `json:"project_id"`
}

type GoogleProvider struct {
	Http HTTPClient
}

func (g GoogleProvider) Name() string {
	return "google cloud"
}

func (g GoogleProvider) DiscoverApplications(info provider.IntegrationInfo) (apps []provider.ApplicationInfo) {
	key := info.Key
	foundCredentials := g.Credentials(key)
	if strings.EqualFold(info.Name, g.Name()) {
		if g.Http == nil {
			g.Http, _ = g.HttpClient(key) // todo - for testing, might be a better way?
		}
		googleClient := GoogleClient{g.Http, foundCredentials.ProjectId}
		found, _ := googleClient.GetBackendApplications()
		apps = append(apps, found...)
	}
	return apps
}

///

func (g GoogleProvider) Credentials(key []byte) credentials {
	var foundCredentials credentials
	_ = json.NewDecoder(bytes.NewReader(key)).Decode(&foundCredentials)
	return foundCredentials
}

func (g GoogleProvider) HttpClient(key []byte) (HTTPClient, error) {
	var opts []option.ClientOption
	opt := option.WithCredentialsJSON(key)
	opts = append([]option.ClientOption{option.WithScopes("https://www.googleapis.com/auth/cloud-platform")}, opt)
	client, _, err := http.NewClient(context.Background(), opts...)
	if err != nil {
		return nil, err
	}
	return client, nil
}
