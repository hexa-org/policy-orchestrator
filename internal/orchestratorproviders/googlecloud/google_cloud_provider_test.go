package googlecloud_test

import (
	"github.com/hexa-org/policy-orchestrator/internal/orchestrator"
	"github.com/hexa-org/policy-orchestrator/internal/orchestratorproviders/googlecloud"
	"github.com/hexa-org/policy-orchestrator/internal/policysupport"
	"testing"

	"github.com/hexa-org/policy-orchestrator/pkg/testsupport"
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
	m := testsupport.NewMockHTTPClient()
	m.ResponseBody["https://compute.googleapis.com/compute/v1/projects/google-cloud-project-id/global/backendServices"] = backendAppsJSON
	m.ResponseBody["https://appengine.googleapis.com/v1/apps/google-cloud-project-id"] = appEngineAppsJSON
	p := &googlecloud.GoogleProvider{HttpClientOverride: m}
	info := orchestrator.IntegrationInfo{Name: "google_cloud", Key: projectJSON}

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
	info := orchestrator.IntegrationInfo{Name: "google_cloud", Key: projectJSON}

	applications, err := p.DiscoverApplications(info)

	assert.NoError(t, err)
	assert.Equal(t, 4, len(applications))
	assert.Equal(t, "google_cloud", p.Name())
}

func TestGoogleProvider_DiscoverApplications_emptyResponse(t *testing.T) {
	m := testsupport.NewMockHTTPClient()
	p := &googlecloud.GoogleProvider{HttpClientOverride: m}
	info := orchestrator.IntegrationInfo{Name: "not google_cloud", Key: []byte("aKey")}

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
	info := orchestrator.IntegrationInfo{Name: "google_cloud", Key: projectJSON}

	infos, _ := p.GetPolicyInfo(info, orchestrator.ApplicationInfo{ObjectID: "k8sObjectId", Name: "k8sName"})

	assert.Equal(t, 2, len(infos))
}

func TestGoogleProvider_SetPolicy(t *testing.T) {
	policy := policysupport.PolicyInfo{
		Meta: policysupport.MetaInfo{Version: "aVersion"}, Actions: []policysupport.ActionInfo{{"anAction"}}, Subject: policysupport.SubjectInfo{Members: []string{"aUser"}}, Object: policysupport.ObjectInfo{
			ResourceID: "anObjectId",
		},
	}
	m := testsupport.NewMockHTTPClient()

	p := googlecloud.GoogleProvider{HttpClientOverride: m}
	info := orchestrator.IntegrationInfo{Name: "google_cloud", Key: []byte("aKey")}
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
	m := testsupport.NewMockHTTPClient()

	p := googlecloud.GoogleProvider{HttpClientOverride: m}
	info := orchestrator.IntegrationInfo{Name: "not google_cloud", Key: []byte("aKey")}

	status, err := p.SetPolicyInfo(info, orchestrator.ApplicationInfo{}, []policysupport.PolicyInfo{missingMeta})
	assert.Equal(t, 500, status)
	assert.Error(t, err)

	status, err = p.SetPolicyInfo(info, orchestrator.ApplicationInfo{ObjectID: "anObjectId"}, []policysupport.PolicyInfo{missingMeta})
	assert.Equal(t, 500, status)
	assert.Error(t, err)
}
