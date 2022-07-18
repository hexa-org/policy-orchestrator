package admin_test

import (
	"github.com/hexa-org/policy-orchestrator/pkg/admin"
	"github.com/stretchr/testify/mock"
)

type MockClient struct {
	mock.Mock
	Name     string
	Provider string
	Key      []byte
	Errs     map[string]error

	DesiredApplication admin.Application
	DesiredPolicies    []admin.Policy
}

func (m *MockClient) Integrations(url string) ([]admin.Integration, error) {
	integration := admin.Integration{ID: "anId", Name: "aName", Provider: "google_cloud", Key: []byte("aKey")}
	return []admin.Integration{integration}, m.Errs[url]
}

func (m *MockClient) CreateIntegration(url string, name string, provider string, key []byte) error {
	args := m.Called(url)
	m.Name = name
	m.Provider = provider
	m.Key = key
	if len(args) > 0 {
		return args.Error(0)
	}
	return m.Errs[url]
}

func (m *MockClient) DeleteIntegration(url string) error {
	args := m.Called(url)
	if len(args) > 0 {
		return args.Error(0)
	}
	return m.Errs[url]
}

func (m *MockClient) Applications(url string) ([]admin.Application, error) {
	return []admin.Application{m.DesiredApplication}, m.Errs[url]
}

func (m *MockClient) Application(url string) (admin.Application, error) {
	return m.DesiredApplication, m.Errs[url]
}

func (m *MockClient) GetPolicies(url string) ([]admin.Policy, string, error) {
	return m.DesiredPolicies, "", m.Errs[url]
}

func (m *MockClient) SetPolicies(url string, _ string) error {
	return m.Errs[url]
}

func (m *MockClient) Health(_ string) (string, error) {
	return "{\"status\":\"pass\"}", nil
}
