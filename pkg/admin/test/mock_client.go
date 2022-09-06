package admin_test

import (
	"fmt"
	"github.com/hexa-org/policy-orchestrator/pkg/admin"
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

	DesiredApplication admin.Application
	DesiredPolicies    []admin.Policy
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

func (m *MockClient) Applications() ([]admin.Application, error) {
	url := fmt.Sprintf("%v/applications", m.Url)
	return []admin.Application{m.DesiredApplication}, m.Errs[url]
}

func (m *MockClient) Application(id string) (admin.Application, error) {
	url := fmt.Sprintf("%v/applications/%s", m.Url, id)
	return m.DesiredApplication, m.Errs[url]
}

func (m *MockClient) GetPolicies(id string) ([]admin.Policy, string, error) {
	url := fmt.Sprintf("%v/applications/%s/policies", m.Url, id)
	return m.DesiredPolicies, "", m.Errs[url]
}

func (m *MockClient) SetPolicies(id string, _ string) error {
	url := fmt.Sprintf("%v/applications/%s/policies", m.Url, id)
	return m.Errs[url]
}

func (m *MockClient) Health() (string, error) {
	return m.Status, nil
}
