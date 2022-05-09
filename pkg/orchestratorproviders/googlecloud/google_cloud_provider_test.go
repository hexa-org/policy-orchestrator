package googlecloud_test

import (
	"github.com/hexa-org/policy-orchestrator/pkg/orchestrator"
	"github.com/hexa-org/policy-orchestrator/pkg/orchestratorproviders/googlecloud"
	"github.com/hexa-org/policy-orchestrator/pkg/orchestratorproviders/googlecloud/test"
	"github.com/hexa-org/policy-orchestrator/pkg/policysupport"
	"github.com/stretchr/testify/assert"
	"io/ioutil"
	"path/filepath"
	"runtime"
	"testing"
)

func TestGoogleProvider_BadClientKey(t *testing.T) {
	p := googlecloud.GoogleProvider{}
	info := orchestrator.IntegrationInfo{Name: "google_cloud", Key: []byte("aKey")}

	_, discoverErr := p.DiscoverApplications(info)
	assert.Error(t, discoverErr)

	_, getErr := p.GetPolicyInfo(info, orchestrator.ApplicationInfo{})
	assert.Error(t, getErr)

	setErr := p.SetPolicyInfo(info, orchestrator.ApplicationInfo{}, []policysupport.PolicyInfo{})
	assert.Error(t, setErr)
}

func TestGoogleProvider_DiscoverApplications(t *testing.T) {
	m := google_cloud_test.NewMockClient()
	m.ResponseBody["compute"] = google_cloud_test.Resource("backends.json")
	m.ResponseBody["appengine"] = google_cloud_test.Resource("appengine.json")
	p := &googlecloud.GoogleProvider{HttpClientOverride: m}

	info := orchestrator.IntegrationInfo{Name: "google_cloud", Key: []byte("aKey")}
	applications, _ := p.DiscoverApplications(info)
	assert.Equal(t, 3, len(applications))
	assert.Equal(t, "google_cloud", p.Name())
}

func TestGoogleProvider_DiscoverApplications_ignoresProviderCase(t *testing.T) {
	m := google_cloud_test.NewMockClient()
	m.ResponseBody["compute"] = google_cloud_test.Resource("backends.json")
	m.ResponseBody["appengine"] = google_cloud_test.Resource("appengine.json")
	p := &googlecloud.GoogleProvider{HttpClientOverride: m}

	info := orchestrator.IntegrationInfo{Name: "Google_Cloud", Key: []byte("aKey")}
	applications, _ := p.DiscoverApplications(info)
	assert.Equal(t, 3, len(applications))
	assert.Equal(t, "google_cloud", p.Name())
}

func TestGoogleProvider_DiscoverApplications_emptyResponse(t *testing.T) {
	m := google_cloud_test.NewMockClient()
	p := &googlecloud.GoogleProvider{HttpClientOverride: m}

	info := orchestrator.IntegrationInfo{Name: "not google_cloud", Key: []byte("aKey")}
	applications, _ := p.DiscoverApplications(info)
	assert.Equal(t, 0, len(applications))
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
	client, _ := p.NewHttpClient(key)
	assert.NotNil(t, client)
}

func TestGoogleProvider_HttpClient_withBadKey(t *testing.T) {
	p := googlecloud.GoogleProvider{}
	_, err := p.NewHttpClient([]byte(""))
	assert.Error(t, err)
}

func TestGoogleProvider_GetPolicy(t *testing.T) {
	m := google_cloud_test.NewMockClient()
	m.ResponseBody["compute"] = google_cloud_test.Resource("policy.json")

	p := googlecloud.GoogleProvider{HttpClientOverride: m}
	info := orchestrator.IntegrationInfo{Name: "not google_cloud", Key: []byte("aKey")}
	infos, _ := p.GetPolicyInfo(info, orchestrator.ApplicationInfo{ObjectID: "anObjectId"})
	assert.Equal(t, 2, len(infos))
}

func TestGoogleProvider_SetPolicy(t *testing.T) {
	policy := policysupport.PolicyInfo{
		Version: "aVersion", Action: "anAction", Subject: policysupport.SubjectInfo{AuthenticatedUsers: []string{"aUser"}}, Object: policysupport.ObjectInfo{Resources: []string{"/"}},
	}
	m := google_cloud_test.NewMockClient()

	p := googlecloud.GoogleProvider{HttpClientOverride: m}
	info := orchestrator.IntegrationInfo{Name: "not google_cloud", Key: []byte("aKey")}
	err := p.SetPolicyInfo(info, orchestrator.ApplicationInfo{ObjectID: "anObjectId"}, []policysupport.PolicyInfo{policy})
	assert.NoError(t, err)
}
