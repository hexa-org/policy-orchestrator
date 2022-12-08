package googlecloud_test

import (
	"errors"
	"github.com/hexa-org/policy-orchestrator/internal/orchestratorproviders/googlecloud"
	"github.com/hexa-org/policy-orchestrator/internal/policysupport"
	"testing"

	"github.com/hexa-org/policy-orchestrator/pkg/testsupport"
	"github.com/stretchr/testify/assert"
)

func TestGoogleClient_GetAppEngineApplications(t *testing.T) {
	m := testsupport.NewMockHTTPClient()
	m.ResponseBody["https://appengine.googleapis.com/v1/apps/projectID"] = appEngineAppsJSON
	client := googlecloud.GoogleClient{ProjectId: "projectID", HttpClient: m}

	applications, _ := client.GetAppEngineApplications()

	assert.Equal(t, 1, len(applications))
	assert.Equal(t, "hexa-demo", applications[0].ObjectID)
	assert.Equal(t, "apps/hexa-demo", applications[0].Name)
	assert.Equal(t, "hexa-demo.uc.r.appspot.com", applications[0].Description)
	assert.Equal(t, "AppEngine", applications[0].Service)
}

func TestGoogleClient_GetAppEngineApplications_when_404(t *testing.T) {
	m := testsupport.NewMockHTTPClient()
	m.StatusCode = 404
	client := googlecloud.GoogleClient{HttpClient: m}

	applications, _ := client.GetAppEngineApplications()

	assert.Equal(t, 0, len(applications))
}

func TestClient_GetAppEngineApplications_withRequestError(t *testing.T) {
	m := testsupport.NewMockHTTPClient()
	m.Err = errors.New("oops")

	client := googlecloud.GoogleClient{HttpClient: m}

	_, err := client.GetAppEngineApplications()
	assert.Error(t, err)
}

func TestClient_GetAppEngineApplications_withBadJson(t *testing.T) {
	m := testsupport.NewMockHTTPClient()
	m.ResponseBody["compute"] = []byte("-")
	client := googlecloud.GoogleClient{HttpClient: m}

	_, err := client.GetAppEngineApplications()

	assert.Error(t, err)
}

func TestClient_GetBackendApplications(t *testing.T) {
	m := testsupport.NewMockHTTPClient()
	m.ResponseBody["https://compute.googleapis.com/compute/v1/projects/projectID/global/backendServices"] = backendAppsJSON
	client := googlecloud.GoogleClient{ProjectId: "projectID", HttpClient: m}

	applications, _ := client.GetBackendApplications()

	assert.Equal(t, 3, len(applications))
	assert.Equal(t, "k8s1-aName", applications[0].Name)
	assert.Equal(t, "k8s1-anotherName", applications[1].Name)
	assert.Equal(t, "cloud-run-app", applications[2].Name)
	assert.Equal(t, "Kubernetes", applications[0].Service)
	assert.Equal(t, "Kubernetes", applications[1].Service)
	assert.Equal(t, "Cloud Run", applications[2].Service)
}

func TestClient_GetBackendApplications_withRequestError(t *testing.T) {
	m := testsupport.NewMockHTTPClient()
	m.Err = errors.New("oops")
	client := googlecloud.GoogleClient{HttpClient: m}

	_, err := client.GetBackendApplications()

	assert.Error(t, err)
}

func TestClient_GetBackendApplications_withBadJson(t *testing.T) {
	m := testsupport.NewMockHTTPClient()
	m.ResponseBody["compute"] = []byte("-")
	client := googlecloud.GoogleClient{HttpClient: m}

	_, err := client.GetBackendApplications()

	assert.Error(t, err)
}

func TestGoogleClient_GetAppEnginePolicies(t *testing.T) {
	m := testsupport.NewMockHTTPClient()
	m.ResponseBody["https://iap.googleapis.com/v1/projects/appengineproject/iap_web/appengine-appEngineObjectId/services/default:getIamPolicy"] = policyJSON
	client := googlecloud.GoogleClient{HttpClient: m, ProjectId: "appengineproject"}
	expectedUsers := []string{
		"user:phil@example.com",
		"group:admins@example.com",
		"domain:google.com",
		"serviceAccount:my-project-id@appspot.gserviceaccount.com",
	}

	infos, _ := client.GetBackendPolicy("apps/EngineName", "appEngineObjectId")

	assert.Equal(t, 2, len(infos))
	assert.Equal(t, expectedUsers, infos[0].Subject.Members)
}

func TestGoogleClient_GetBackendPolicies(t *testing.T) {
	m := testsupport.NewMockHTTPClient()
	m.ResponseBody["https://iap.googleapis.com/v1/projects/k8sproject/iap_web/compute/services/k8sObjectId:getIamPolicy"] = policyJSON
	client := googlecloud.GoogleClient{HttpClient: m, ProjectId: "k8sproject"}
	expectedUsers := []string{
		"user:phil@example.com",
		"group:admins@example.com",
		"domain:google.com",
		"serviceAccount:my-project-id@appspot.gserviceaccount.com",
	}

	infos, _ := client.GetBackendPolicy("k8sName", "k8sObjectId")

	assert.Equal(t, 2, len(infos))
	assert.Equal(t, expectedUsers, infos[0].Subject.Members)
}

func TestGoogleClient_GetBackendPolicies_withRequestError(t *testing.T) {
	m := testsupport.NewMockHTTPClient()
	m.Err = errors.New("oops")
	client := googlecloud.GoogleClient{HttpClient: m}

	_, err := client.GetBackendPolicy("k8sName", "anObjectId")

	assert.Error(t, err)
}

func TestGoogleClient_GetBackendPolicies_withBadJson(t *testing.T) {
	m := testsupport.NewMockHTTPClient()
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
	m := testsupport.NewMockHTTPClient()
	client := googlecloud.GoogleClient{HttpClient: m, ProjectId: "appengineproject"}

	err := client.SetBackendPolicy("apps/EngineName", "anObjectId", policy)

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
	m := testsupport.NewMockHTTPClient()
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
	m := testsupport.NewMockHTTPClient()
	m.Err = errors.New("oops")
	client := googlecloud.GoogleClient{HttpClient: m}
	err := client.SetBackendPolicy("k8sName", "anObjectId", policy)
	assert.Error(t, err)
}
