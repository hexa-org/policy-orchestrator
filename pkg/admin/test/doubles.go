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
	Err      error
}

func (m *MockClient) Integrations(_ string) ([]admin.Integration, error) {
	integration := admin.Integration{ID: "anId", Name: "aName", Provider: "google_cloud", Key: []byte("aKey")}
	return []admin.Integration{integration}, m.Err
}

func (m *MockClient) CreateIntegration(url string, name string, provider string, key []byte) error {
	args := m.Called(url)
	m.Name = name
	m.Provider = provider
	m.Key = key
	if len(args) > 0 {
		return args.Error(0)
	}
	return m.Err
}

func (m *MockClient) DeleteIntegration(url string) error {
	args := m.Called(url)
	if len(args) > 0 {
		return args.Error(0)
	}
	return m.Err
}

func (m *MockClient) Applications(_ string) ([]admin.Application, error) {
	application := admin.Application{ID: "anId", IntegrationId: "anIntegrationId", ObjectId: "anObjectId", Name: "aName", Description: "aDescription", ProviderName: "google_cloud"} // keep for now - tests proper provider name rendering
	return []admin.Application{application}, m.Err
}

func (m *MockClient) Application(_ string) (admin.Application, error) {
	return admin.Application{ID: "anId", IntegrationId: "anIntegrationId", ObjectId: "anObjectId", Name: "aName", Description: "aDescription"}, m.Err
}

func (m *MockClient) GetPolicies(_ string) ([]admin.Policy, string, error) {
	return []admin.Policy{
		{admin.Meta{Version: "aVersion"}, []admin.Action{{"anAction"}}, admin.Subject{Members: []string{"aUser"}}, admin.Object{ResourceID: "aResourceId", Resources: []string{"/"}}},
		{admin.Meta{Version: "aVersion"}, []admin.Action{{"anotherAction"}}, admin.Subject{Members: []string{"anotherUser"}}, admin.Object{ResourceID: "anotherResourceId", Resources: []string{"/"}}},
	}, "", m.Err
}

func (m *MockClient) SetPolicies(_ string, _ string) error {
	return m.Err
}

func (m *MockClient) Health(_ string) (string, error) {
	return "{\"status\":\"pass\"}", nil
}
