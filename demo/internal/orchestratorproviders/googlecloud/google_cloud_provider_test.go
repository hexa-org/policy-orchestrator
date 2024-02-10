package googlecloud_test

import (
	"testing"

	"github.com/hexa-org/policy-mapper/api/policyprovider"
	"github.com/hexa-org/policy-mapper/pkg/hexapolicy"
	"github.com/hexa-org/policy-orchestrator/demo/internal/orchestratorproviders/googlecloud"
	"github.com/hexa-org/policy-orchestrator/demo/pkg/testsupport"
	"github.com/stretchr/testify/assert"
)

func TestGoogleProvider_BadClientKey(t *testing.T) {
	p := googlecloud.GoogleProvider{}
	info := policyprovider.IntegrationInfo{Name: "google_cloud", Key: []byte("aKey")}

	_, discoverErr := p.DiscoverApplications(info)
	assert.Error(t, discoverErr)

	_, getErr := p.GetPolicyInfo(info, policyprovider.ApplicationInfo{ObjectID: "anObjectId"})
	assert.Error(t, getErr)

	status, setErr := p.SetPolicyInfo(info, policyprovider.ApplicationInfo{ObjectID: "anObjectId"}, []hexapolicy.PolicyInfo{})
	assert.Equal(t, 500, status)
	assert.Error(t, setErr)
}

func TestGoogleProvider_DiscoverApplications(t *testing.T) {
	m := testsupport.NewMockHTTPClient()
	m.ResponseBody["https://compute.googleapis.com/compute/v1/projects/google-cloud-project-id/global/backendServices"] = backendAppsJSON
	m.ResponseBody["https://appengine.googleapis.com/v1/apps/google-cloud-project-id"] = appEngineAppsJSON
	p := &googlecloud.GoogleProvider{HttpClientOverride: m}
	info := policyprovider.IntegrationInfo{Name: "google_cloud", Key: projectJSON}

	applications, err := p.DiscoverApplications(info)

	assert.NoError(t, err)
	assert.Equal(t, 4, len(applications))
	assert.Equal(t, "Kubernetes", applications[0].Service)
	assert.Equal(t, "Kubernetes", applications[1].Service)
	assert.Equal(t, "Cloud Run", applications[2].Service)
	assert.Equal(t, "AppEngine", applications[3].Service)
	assert.Equal(t, "google_cloud", p.Name())
}

func TestGoogleProvider_DiscoverApplications_ignoresProviderCase(t *testing.T) {
	m := testsupport.NewMockHTTPClient()
	m.ResponseBody["https://compute.googleapis.com/compute/v1/projects/google-cloud-project-id/global/backendServices"] = backendAppsJSON
	m.ResponseBody["https://appengine.googleapis.com/v1/apps/google-cloud-project-id"] = appEngineAppsJSON
	p := &googlecloud.GoogleProvider{HttpClientOverride: m}
	info := policyprovider.IntegrationInfo{Name: "google_cloud", Key: projectJSON}

	applications, err := p.DiscoverApplications(info)

	assert.NoError(t, err)
	assert.Equal(t, 4, len(applications))
	assert.Equal(t, "google_cloud", p.Name())
}

func TestGoogleProvider_DiscoverApplications_emptyResponse(t *testing.T) {
	m := testsupport.NewMockHTTPClient()
	p := &googlecloud.GoogleProvider{HttpClientOverride: m}
	info := policyprovider.IntegrationInfo{Name: "not google_cloud", Key: []byte("aKey")}

	applications, _ := p.DiscoverApplications(info)

	assert.Equal(t, 0, len(applications))
}

func TestGoogleProvider_Project(t *testing.T) {
	p := googlecloud.GoogleProvider{}

	assert.Equal(t, "google-cloud-project-id", p.Project(projectJSON))
}

func TestGoogleProvider_HttpClient(t *testing.T) {
	p := googlecloud.GoogleProvider{}

	client, err := p.NewHttpClient(projectJSON)

	assert.NotNil(t, client)
	assert.NoError(t, err)
}

func TestGoogleProvider_HttpClient_withBadKey(t *testing.T) {
	p := googlecloud.GoogleProvider{}

	_, err := p.NewHttpClient([]byte(""))

	assert.Error(t, err)
}

func TestGoogleProvider_GetPolicy(t *testing.T) {
	m := testsupport.NewMockHTTPClient()
	m.ResponseBody["https://iap.googleapis.com/v1/projects/google-cloud-project-id/iap_web/compute/services/k8sObjectId:getIamPolicy"] = policyJSON
	p := googlecloud.GoogleProvider{HttpClientOverride: m}
	info := policyprovider.IntegrationInfo{Name: "google_cloud", Key: projectJSON}

	infos, _ := p.GetPolicyInfo(info, policyprovider.ApplicationInfo{ObjectID: "k8sObjectId", Name: "k8sName"})

	assert.Equal(t, 2, len(infos))
}

func TestGoogleProvider_SetPolicy(t *testing.T) {
	policy := hexapolicy.PolicyInfo{
		Meta: hexapolicy.MetaInfo{Version: "aVersion"}, Actions: []hexapolicy.ActionInfo{{"anAction"}}, Subject: hexapolicy.SubjectInfo{Members: []string{"aUser"}}, Object: hexapolicy.ObjectInfo{
			ResourceID: "anObjectId",
		},
	}
	m := testsupport.NewMockHTTPClient()

	p := googlecloud.GoogleProvider{HttpClientOverride: m}
	info := policyprovider.IntegrationInfo{Name: "google_cloud", Key: []byte("aKey")}
	status, err := p.SetPolicyInfo(info, policyprovider.ApplicationInfo{ObjectID: "anObjectId"}, []hexapolicy.PolicyInfo{policy})
	assert.Equal(t, 201, status)
	assert.NoError(t, err)
}

func TestGoogleProvider_SetPolicy_withInvalidArguments(t *testing.T) {
	missingMeta := hexapolicy.PolicyInfo{
		Actions: []hexapolicy.ActionInfo{{"anAction"}}, Subject: hexapolicy.SubjectInfo{Members: []string{"aUser"}}, Object: hexapolicy.ObjectInfo{
			ResourceID: "anObjectId",
		},
	}
	m := testsupport.NewMockHTTPClient()

	p := googlecloud.GoogleProvider{HttpClientOverride: m}
	info := policyprovider.IntegrationInfo{Name: "not google_cloud", Key: []byte("aKey")}

	status, err := p.SetPolicyInfo(info, policyprovider.ApplicationInfo{}, []hexapolicy.PolicyInfo{missingMeta})
	assert.Equal(t, 500, status)
	assert.Error(t, err)

	status, err = p.SetPolicyInfo(info, policyprovider.ApplicationInfo{ObjectID: "anObjectId"}, []hexapolicy.PolicyInfo{missingMeta})
	assert.Equal(t, 500, status)
	assert.Error(t, err)
}
