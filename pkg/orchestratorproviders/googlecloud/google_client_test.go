package googlecloud_test

import (
	"errors"
	"testing"

	"github.com/hexa-org/policy-orchestrator/pkg/orchestratorproviders/googlecloud"
	"github.com/hexa-org/policy-orchestrator/pkg/orchestratorproviders/googlecloud/test"
	"github.com/hexa-org/policy-orchestrator/pkg/policysupport"
	"github.com/stretchr/testify/assert"
)

func TestGoogleClient_GetAppEngineApplications(t *testing.T) {
	m := google_cloud_test.NewMockClient()
	m.ResponseBody["appengine"] = google_cloud_test.Resource("appengine.json")
	client := googlecloud.GoogleClient{HttpClient: m}
	applications, _ := client.GetAppEngineApplications()

	assert.Equal(t, 1, len(applications))
	assert.Equal(t, "hexa-demo", applications[0].ObjectID)
	assert.Equal(t, "apps/hexa-demo", applications[0].Name)
	assert.Equal(t, "hexa-demo.uc.r.appspot.com", applications[0].Description)
	assert.Equal(t, "AppEngine", applications[0].Service)
}

func TestGoogleClient_GetAppEngineApplications_when_404(t *testing.T) {
	m := google_cloud_test.NewMockClient()
	m.StatusCode = 404
	client := googlecloud.GoogleClient{HttpClient: m}
	applications, _ := client.GetAppEngineApplications()

	assert.Equal(t, 0, len(applications))
}

func TestClient_GetAppEngineApplications_withRequestError(t *testing.T) {
	m := google_cloud_test.NewMockClient()
	m.Err = errors.New("oops")

	client := googlecloud.GoogleClient{HttpClient: m}
	_, err := client.GetAppEngineApplications()
	assert.Error(t, err)
}

func TestClient_GetAppEngineApplications_withBadJson(t *testing.T) {
	m := google_cloud_test.NewMockClient()
	m.ResponseBody["compute"] = []byte("-")
	client := googlecloud.GoogleClient{HttpClient: m}
	_, err := client.GetAppEngineApplications()
	assert.Error(t, err)
}

func TestClient_GetBackendApplications(t *testing.T) {
	m := google_cloud_test.NewMockClient()
	m.ResponseBody["compute"] = google_cloud_test.Resource("backends.json")
	client := googlecloud.GoogleClient{HttpClient: m}
	applications, _ := client.GetBackendApplications()

	assert.Equal(t, 2, len(applications))
	assert.Equal(t, "k8s1-aName", applications[0].Name)
	assert.Equal(t, "k8s1-anotherName", applications[1].Name)
	assert.Equal(t, "Kubernetes", applications[0].Service)
}

func TestClient_GetBackendApplications_withRequestError(t *testing.T) {
	m := google_cloud_test.NewMockClient()
	m.Err = errors.New("oops")

	client := googlecloud.GoogleClient{HttpClient: m}
	_, err := client.GetBackendApplications()
	assert.Error(t, err)
}

func TestClient_GetBackendApplications_withBadJson(t *testing.T) {
	m := google_cloud_test.NewMockClient()
	m.ResponseBody["compute"] = []byte("-")
	client := googlecloud.GoogleClient{HttpClient: m}
	_, err := client.GetBackendApplications()
	assert.Error(t, err)
}

func TestGoogleClient_GetAppEnginePolicies(t *testing.T) {
	m := google_cloud_test.NewMockClient()
	m.ResponseBody["appengine"] = google_cloud_test.Resource("policy.json")
	client := googlecloud.GoogleClient{HttpClient: m, ProjectId: "appengineproject"}
	infos, _ := client.GetBackendPolicy("appEngineName", "appEngineObjectId")

	expectedUsers := []string{
		"user:phil@example.com",
		"group:admins@example.com",
		"domain:google.com",
		"serviceAccount:my-project-id@appspot.gserviceaccount.com",
	}
	assert.Equal(t, 2, len(infos))
	assert.Equal(t, expectedUsers, infos[0].Subject.Members)
	assert.Equal(t, "https://iap.googleapis.com/v1/projects/appengineproject/iap_web/appengine-appEngineObjectId/services/default:getIamPolicy", m.Url)
}

func TestGoogleClient_GetBackendPolicies(t *testing.T) {
	m := google_cloud_test.NewMockClient()
	m.ResponseBody["compute"] = google_cloud_test.Resource("policy.json")
	client := googlecloud.GoogleClient{HttpClient: m, ProjectId: "k8sproject"}
	infos, _ := client.GetBackendPolicy("k8sName", "k8sObjectId")

	expectedUsers := []string{
		"user:phil@example.com",
		"group:admins@example.com",
		"domain:google.com",
		"serviceAccount:my-project-id@appspot.gserviceaccount.com",
	}
	assert.Equal(t, 2, len(infos))
	assert.Equal(t, expectedUsers, infos[0].Subject.Members)
	assert.Equal(t, "https://iap.googleapis.com/v1/projects/k8sproject/iap_web/compute/services/k8sObjectId:getIamPolicy", m.Url)
}

func TestGoogleClient_GetBackendPolicies_withRequestError(t *testing.T) {
	m := google_cloud_test.NewMockClient()
	m.Err = errors.New("oops")

	client := googlecloud.GoogleClient{HttpClient: m}
	_, err := client.GetBackendPolicy("k8sName", "anObjectId")
	assert.Error(t, err)
}

func TestGoogleClient_GetBackendPolicies_withBadJson(t *testing.T) {
	m := google_cloud_test.NewMockClient()
	m.ResponseBody["compute"] = []byte("-")
	client := googlecloud.GoogleClient{HttpClient: m}
	_, err := client.GetBackendPolicy("k8sName", "anObjectId")
	assert.Error(t, err)
}

func TestGoogleClient_SetAppEnginePolicies(t *testing.T) {
	policy := policysupport.PolicyInfo{
		Meta: policysupport.MetaInfo{Version: "aVersion"}, Actions: []policysupport.ActionInfo{{"roles/iap.httpsResourceAccessor"}}, Subject: policysupport.SubjectInfo{Members: []string{"aUser"}}, Object: policysupport.ObjectInfo{
			ResourceID: "anObjectId",
		},
	}
	m := google_cloud_test.NewMockClient()
	client := googlecloud.GoogleClient{HttpClient: m, ProjectId: "appengineproject"}
	err := client.SetBackendPolicy("appEngineName", "anObjectId", policy)
	assert.NoError(t, err)
	assert.Equal(t, "{\"policy\":{\"bindings\":[{\"role\":\"roles/iap.httpsResourceAccessor\",\"members\":[\"aUser\"]}]}}\n", string(m.RequestBody))
	assert.Equal(t, "https://iap.googleapis.com/v1/projects/appengineproject/iap_web/appengine-anObjectId/services/default:setIamPolicy", m.Url)
}

func TestGoogleClient_SetBackendPolicies(t *testing.T) {
	policy := policysupport.PolicyInfo{
		Meta: policysupport.MetaInfo{Version: "aVersion"}, Actions: []policysupport.ActionInfo{{"gcp:roles/iap.httpsResourceAccessor"}}, Subject: policysupport.SubjectInfo{Members: []string{"aUser"}}, Object: policysupport.ObjectInfo{
			ResourceID: "anObjectId",
		},
	}
	m := google_cloud_test.NewMockClient()
	client := googlecloud.GoogleClient{HttpClient: m, ProjectId: "k8sproject"}
	err := client.SetBackendPolicy("k8sName", "anObjectId", policy)
	assert.NoError(t, err)
	assert.Equal(t, "{\"policy\":{\"bindings\":[{\"role\":\"roles/iap.httpsResourceAccessor\",\"members\":[\"aUser\"]}]}}\n", string(m.RequestBody))
	assert.Equal(t, "https://iap.googleapis.com/v1/projects/k8sproject/iap_web/compute/services/anObjectId:setIamPolicy", m.Url)
}

func TestGoogleClient_SetBackendPolicies_withRequestError(t *testing.T) {
	policy := policysupport.PolicyInfo{
		Meta: policysupport.MetaInfo{Version: "aVersion"}, Actions: []policysupport.ActionInfo{{"gcp:roles/iap.httpsResourceAccessor"}}, Subject: policysupport.SubjectInfo{Members: []string{"aUser"}}, Object: policysupport.ObjectInfo{
			ResourceID: "anObjectId",
		},
	}
	m := google_cloud_test.NewMockClient()
	m.Err = errors.New("oops")
	client := googlecloud.GoogleClient{HttpClient: m}
	err := client.SetBackendPolicy("k8sName", "anObjectId", policy)
	assert.Error(t, err)
}
