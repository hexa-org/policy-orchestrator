package google_cloud_test

import (
	"errors"
	"github.com/hexa-org/policy-orchestrator/pkg/providers/google_cloud"
	"github.com/hexa-org/policy-orchestrator/pkg/providers/google_cloud/test"
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestClient_GetBackendApplications(t *testing.T) {
	m := new(google_cloud_test.MockClient)
	client := google_cloud.GoogleClient{HttpClient: m}
	applications, _ := client.GetBackendApplications()

	assert.Equal(t, 2, len(applications))
	assert.Equal(t, "k8s1-aName", applications[0].Name)
	assert.Equal(t, "k8s1-anotherName", applications[1].Name)
}

func TestClient_GetBackendApplications_with_error(t *testing.T) {
	m := new(google_cloud_test.MockClient)
	m.Err = errors.New("oops")

	client := google_cloud.GoogleClient{HttpClient: m}
	_, err := client.GetBackendApplications()
	assert.Error(t, err)
}

func TestClient_GetBackendApplications_with_bad_json(t *testing.T) {
	m := new(google_cloud_test.MockClient)
	m.Json = "-"

	client := google_cloud.GoogleClient{HttpClient: m}
	_, err := client.GetBackendApplications()
	assert.Error(t, err)
}
