package googlecloud_test

import (
	"github.com/hexa-org/policy-orchestrator/pkg/orchestrator/provider"
	"github.com/hexa-org/policy-orchestrator/pkg/providers/googlecloud"
	"github.com/hexa-org/policy-orchestrator/pkg/providers/googlecloud/test"
	"github.com/stretchr/testify/assert"
	"io/ioutil"
	"path/filepath"
	"runtime"
	"testing"
)

func TestDiscovery(t *testing.T) {
	m := new(google_cloud_test.MockClient)
	providers := []provider.Provider{googlecloud.GoogleProvider{Http: m}}

	for _, p := range providers {
		info := provider.IntegrationInfo{Name: "google cloud", Key: []byte("aKey")}
		assert.Equal(t, 2, len(p.DiscoverApplications(info)))
		assert.Equal(t, "google cloud", p.Name())
	}
}

func TestDiscovery_ignores_case(t *testing.T) {
	m := new(google_cloud_test.MockClient)
	providers := []provider.Provider{googlecloud.GoogleProvider{Http: m}}

	for _, p := range providers {
		info := provider.IntegrationInfo{Name: "Google Cloud", Key: []byte("aKey")}
		assert.Equal(t, 2, len(p.DiscoverApplications(info)))
		assert.Equal(t, "google cloud", p.Name())
	}
}

func TestNoDiscovery(t *testing.T) {
	m := new(google_cloud_test.MockClient)
	providers := []provider.Provider{googlecloud.GoogleProvider{Http: m}}

	for _, p := range providers {
		info := provider.IntegrationInfo{Name: "not google cloud", Key: []byte("aKey")}
		assert.Equal(t, 0, len(p.DiscoverApplications(info)))
	}
}

func TestGoogleProvider_DetermineProjectId(t *testing.T) {
	_, file, _, _ := runtime.Caller(0)
	jsonFile := filepath.Join(file, "./../test/project.json")
	key, _ := ioutil.ReadFile(jsonFile)

	p := googlecloud.GoogleProvider{}
	foundCredentials := p.Credentials(key)
	assert.Equal(t, "google-cloud-project-id", foundCredentials.ProjectId)
}

func TestClient(t *testing.T) {
	_, file, _, _ := runtime.Caller(0)
	jsonFile := filepath.Join(file, "./../test/project.json")
	key, _ := ioutil.ReadFile(jsonFile)

	p := googlecloud.GoogleProvider{}
	client, _ := p.HttpClient(key)
	assert.NotNil(t, client)

	_, err := p.HttpClient([]byte(""))
	assert.Error(t, err)
}
