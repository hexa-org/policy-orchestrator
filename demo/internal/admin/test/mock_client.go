package admin_test

import (
	"fmt"
	"net/http"

	"github.com/hexa-org/policy-mapper/pkg/hexapolicy"
	"github.com/hexa-org/policy-orchestrator/demo/internal/admin"

	"github.com/stretchr/testify/mock"
)

type MockClient struct {
	mock.Mock
	Name     string
	Provider string
	Key      []byte
	Errs     map[string]error
	Status   string
	Url      string

	DesiredApplications []admin.Application
	DesiredPolicies     []hexapolicy.PolicyInfo
}

// GetHttpClient used mainly for testing
func (m *MockClient) GetHttpClient() admin.HTTPClient {
	return &http.Client{}
}

func (m *MockClient) Health() (string, error) {
	return m.Status, nil
}

func (m *MockClient) Integrations() ([]admin.Integration, error) {
	integration := admin.Integration{ID: "anId", Name: "aName", Provider: "google_cloud", Key: []byte("aKey")}
	url := fmt.Sprintf("%v/integrations", m.Url)
	return []admin.Integration{integration}, m.Errs[url]
}

func (m *MockClient) CreateIntegration(name string, provider string, key []byte) error {
	url := fmt.Sprintf("%v/integrations", m.Url)
	args := m.Called(url)
	m.Name = name
	m.Provider = provider
	m.Key = key
	if len(args) > 0 {
		return args.Error(0)
	}
	return m.Errs[url]
}

func (m *MockClient) DeleteIntegration(id string) error {
	url := fmt.Sprintf("%v/integrations/%s", m.Url, id)
	args := m.Called(url)
	if len(args) > 0 {
		return args.Error(0)
	}
	return m.Errs[url]
}

func (m *MockClient) Applications(bool) ([]admin.Application, error) {
	url := fmt.Sprintf("%v/applications", m.Url)
	return m.DesiredApplications, m.Errs[url]
}

func (m *MockClient) Application(id string) (admin.Application, error) {
	url := fmt.Sprintf("%v/applications/%s", m.Url, id)
	if len(m.DesiredApplications) == 0 {
		return admin.Application{}, m.Errs[url]
	}
	return m.DesiredApplications[0], m.Errs[url]
}

func (m *MockClient) GetPolicies(id string) ([]hexapolicy.PolicyInfo, string, error) {
	url := fmt.Sprintf("%v/applications/%s/policies", m.Url, id)
	return m.DesiredPolicies, "{\"policies\":[]}", m.Errs[url]
}

func (m *MockClient) SetPolicies(id string, _ string) error {
	url := fmt.Sprintf("%v/applications/%s/policies", m.Url, id)
	return m.Errs[url]
}

func (m *MockClient) Orchestration(_ string, _ string) error {
	url := fmt.Sprintf("%v/orchestration", m.Url)
	return m.Errs[url]
}
