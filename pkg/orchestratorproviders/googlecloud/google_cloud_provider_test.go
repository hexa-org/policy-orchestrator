package googlecloud_test

import (
	"io/ioutil"
	"path/filepath"
	"runtime"
	"testing"

	"github.com/hexa-org/policy-orchestrator/pkg/orchestrator"
	"github.com/hexa-org/policy-orchestrator/pkg/orchestratorproviders/googlecloud"
	"github.com/hexa-org/policy-orchestrator/pkg/orchestratorproviders/googlecloud/test"
	"github.com/hexa-org/policy-orchestrator/pkg/policysupport"
	"github.com/stretchr/testify/assert"
)

func TestGoogleProvider_BadClientKey(t *testing.T) {
	p := googlecloud.GoogleProvider{}
	info := orchestrator.IntegrationInfo{Name: "google_cloud", Key: []byte("aKey")}

	_, discoverErr := p.DiscoverApplications(info)
	assert.Error(t, discoverErr)

	_, getErr := p.GetPolicyInfo(info, orchestrator.ApplicationInfo{ObjectID: "anObjectId"})
	assert.Error(t, getErr)

	status, setErr := p.SetPolicyInfo(info, orchestrator.ApplicationInfo{ObjectID: "anObjectId"}, []policysupport.PolicyInfo{})
	assert.Equal(t, 500, status)
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
	infos, _ := p.GetPolicyInfo(info, orchestrator.ApplicationInfo{ObjectID: "k8sObjectId", Name: "k8sName"})
	assert.Equal(t, 2, len(infos))
}

func TestGoogleProvider_SetPolicy(t *testing.T) {
	policy := policysupport.PolicyInfo{
		Meta: policysupport.MetaInfo{Version: "aVersion"}, Actions: []policysupport.ActionInfo{{"anAction"}}, Subject: policysupport.SubjectInfo{Members: []string{"aUser"}}, Object: policysupport.ObjectInfo{
			ResourceID: "anObjectId",
		},
	}
	m := google_cloud_test.NewMockClient()

	p := googlecloud.GoogleProvider{HttpClientOverride: m}
	info := orchestrator.IntegrationInfo{Name: "not google_cloud", Key: []byte("aKey")}
	status, err := p.SetPolicyInfo(info, orchestrator.ApplicationInfo{ObjectID: "anObjectId"}, []policysupport.PolicyInfo{policy})
	assert.Equal(t, 201, status)
	assert.NoError(t, err)
}

func TestGoogleProvider_SetPolicy_withInvalidArguments(t *testing.T) {
	missingMeta := policysupport.PolicyInfo{
		Actions: []policysupport.ActionInfo{{"anAction"}}, Subject: policysupport.SubjectInfo{Members: []string{"aUser"}}, Object: policysupport.ObjectInfo{
			ResourceID: "anObjectId",
		},
	}
	m := google_cloud_test.NewMockClient()

	p := googlecloud.GoogleProvider{HttpClientOverride: m}
	info := orchestrator.IntegrationInfo{Name: "not google_cloud", Key: []byte("aKey")}

	status, err := p.SetPolicyInfo(info, orchestrator.ApplicationInfo{}, []policysupport.PolicyInfo{missingMeta})
	assert.Equal(t, 500, status)
	assert.Error(t, err)

	status, err = p.SetPolicyInfo(info, orchestrator.ApplicationInfo{ObjectID: "anObjectId"}, []policysupport.PolicyInfo{missingMeta})
	assert.Equal(t, 500, status)
	assert.Error(t, err)
}
