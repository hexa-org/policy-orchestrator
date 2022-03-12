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

func TestGoogleProvider_DiscoverApplications(t *testing.T) {
	m := new(google_cloud_test.MockClient)
	m.ResponseBody = google_cloud_test.Resource("backends.json")
	providers := []provider.Provider{&googlecloud.GoogleProvider{Http: m}}

	for _, p := range providers {
		info := provider.IntegrationInfo{Name: "google_cloud", Key: []byte("aKey")}
		applications, _ := p.DiscoverApplications(info)
		assert.Equal(t, 2, len(applications))
		assert.Equal(t, "google_cloud", p.Name())
	}
}

func TestGoogleProvider_DiscoverApplications_ignoresProviderCase(t *testing.T) {
	m := new(google_cloud_test.MockClient)
	m.ResponseBody = google_cloud_test.Resource("backends.json")
	providers := []provider.Provider{&googlecloud.GoogleProvider{Http: m}}

	for _, p := range providers {
		info := provider.IntegrationInfo{Name: "Google_Cloud", Key: []byte("aKey")}
		applications, _ := p.DiscoverApplications(info)
		assert.Equal(t, 2, len(applications))
		assert.Equal(t, "google_cloud", p.Name())
	}
}

func TestGoogleProvider_DiscoverApplications_emptyResponse(t *testing.T) {
	m := new(google_cloud_test.MockClient)
	providers := []provider.Provider{&googlecloud.GoogleProvider{Http: m}}

	for _, p := range providers {
		info := provider.IntegrationInfo{Name: "not google_cloud", Key: []byte("aKey")}
		applications, _ := p.DiscoverApplications(info)
		assert.Equal(t, 0, len(applications))
	}
}

func TestGoogleProvider_Project(t *testing.T) {
	_, file, _, _ := runtime.Caller(0)
	jsonFile := filepath.Join(file, "./../test/project.json")
	key, _ := ioutil.ReadFile(jsonFile)

	p := googlecloud.GoogleProvider{}
	assert.Equal(t, "google-cloud-project-id", p.Project(key))
}

func TestGoogleProvider_HttpClient(t *testing.T) {
	_, file, _, _ := runtime.Caller(0)
	jsonFile := filepath.Join(file, "./../test/project.json")
	key, _ := ioutil.ReadFile(jsonFile)

	p := googlecloud.GoogleProvider{}
	client, _ := p.HttpClient(key)
	assert.NotNil(t, client)
}

func TestGoogleProvider_HttpClient_withBadKey(t *testing.T) {
	p := googlecloud.GoogleProvider{}
	_, err := p.HttpClient([]byte(""))
	assert.Error(t, err)
}

func TestGoogleProvider_GetPolicy(t *testing.T) {
	m := new(google_cloud_test.MockClient)
	m.ResponseBody = google_cloud_test.Resource("policy.json")

	p := googlecloud.GoogleProvider{Http: m}
	info := provider.IntegrationInfo{Name: "not google_cloud", Key: []byte("aKey")}
	infos, _ := p.GetPolicyInfo(info, provider.ApplicationInfo{ObjectID: "anObjectId"})
	assert.Equal(t, 2, len(infos))
}

func TestGoogleProvider_SetPolicy(t *testing.T) {
	policy := provider.PolicyInfo{
		Version: "aVersion", Action: "anAction", Subject: provider.SubjectInfo{AuthenticatedUsers: []string{"aUser"}}, Object: provider.ObjectInfo{Resources: []string{"/"}},
	}
	m := new(google_cloud_test.MockClient)

	p := googlecloud.GoogleProvider{Http: m}
	info := provider.IntegrationInfo{Name: "not google_cloud", Key: []byte("aKey")}
	err := p.SetPolicyInfo(info, provider.ApplicationInfo{ObjectID: "anObjectId"}, []provider.PolicyInfo{policy})
	assert.NoError(t, err)
}
