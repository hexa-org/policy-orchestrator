package googlecloud_test

import (
	"errors"
	"github.com/hexa-org/policy-orchestrator/pkg/orchestrator/provider"
	"github.com/hexa-org/policy-orchestrator/pkg/providers/googlecloud"
	"github.com/hexa-org/policy-orchestrator/pkg/providers/googlecloud/test"
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestClient_GetBackendApplications(t *testing.T) {
	m := new(google_cloud_test.MockClient)
	m.Json = google_cloud_test.Resource("backends.json")
	client := googlecloud.GoogleClient{HttpClient: m}
	applications, _ := client.GetBackendApplications()

	assert.Equal(t, 2, len(applications))
	assert.Equal(t, "k8s1-aName", applications[0].Name)
	assert.Equal(t, "k8s1-anotherName", applications[1].Name)
}

func TestClient_GetBackendApplications_withRequestError(t *testing.T) {
	m := new(google_cloud_test.MockClient)
	m.Err = errors.New("oops")

	client := googlecloud.GoogleClient{HttpClient: m}
	_, err := client.GetBackendApplications()
	assert.Error(t, err)
}

func TestClient_GetBackendApplications_withBadJson(t *testing.T) {
	m := new(google_cloud_test.MockClient)
	m.Json = []byte("-")
	client := googlecloud.GoogleClient{HttpClient: m}
	_, err := client.GetBackendApplications()
	assert.Error(t, err)
}

func TestGoogleClient_GetBackendPolicies(t *testing.T) {
	m := new(google_cloud_test.MockClient)
	m.Json = google_cloud_test.Resource("policy.json")
	client := googlecloud.GoogleClient{HttpClient: m}
	infos, _ := client.GetBackendPolicy("anObjectId")

	expectedUsers := []string{
		"user:phil@example.com",
		"group:admins@example.com",
		"domain:google.com",
		"serviceAccount:my-project-id@appspot.gserviceaccount.com",
	}
	assert.Equal(t, 2, len(infos))
	assert.Equal(t, expectedUsers, infos[0].Subject.AuthenticatedUsers)
	assert.Equal(t, []string{"/"}, infos[1].Object.Resources)
}

func TestGoogleClient_GetBackendPolicies_withRequestError(t *testing.T) {
	m := new(google_cloud_test.MockClient)
	m.Err = errors.New("oops")

	client := googlecloud.GoogleClient{HttpClient: m}
	_, err := client.GetBackendPolicy("anObjectId")
	assert.Error(t, err)
}

func TestGoogleClient_GetBackendPolicies_withBadJson(t *testing.T) {
	m := new(google_cloud_test.MockClient)
	m.Json = []byte("-")
	client := googlecloud.GoogleClient{HttpClient: m}
	_, err := client.GetBackendPolicy("anObjectId")
	assert.Error(t, err)
}

func TestGoogleClient_SetBackendPolicies(t *testing.T) {
	policy := provider.PolicyInfo{
		Version: "aVersion", Action: "anAction", Subject: provider.SubjectInfo{AuthenticatedUsers: []string{"aUser"}}, Object: provider.ObjectInfo{Resources: []string{"/"}},
	}
	m := new(google_cloud_test.MockClient)
	client := googlecloud.GoogleClient{HttpClient: m}
	err := client.SetBackendPolicy("anObjectId", policy)
	assert.NoError(t, err)
	assert.Equal(t,"{\"policy\":{\"bindings\":[{\"role\":\"anAction\",\"members\":[\"aUser\"]}]}}\n" , string(m.Json))
}

func TestGoogleClient_SetBackendPolicies_withRequestError(t *testing.T) {
	policy := provider.PolicyInfo{
		Version: "aVersion", Action: "anAction", Subject: provider.SubjectInfo{AuthenticatedUsers: []string{"aUser"}}, Object: provider.ObjectInfo{Resources: []string{"/"}},
	}
	m := new(google_cloud_test.MockClient)
	m.Err = errors.New("oops")
	client := googlecloud.GoogleClient{HttpClient: m}
	err := client.SetBackendPolicy("anObjectId", policy)
	assert.Error(t, err)
}
