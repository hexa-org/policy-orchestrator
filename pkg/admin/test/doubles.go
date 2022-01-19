package admin_test

import (
	"github.com/hexa-org/policy-orchestrator/pkg/admin"
	"github.com/stretchr/testify/mock"
)

type MockClient struct {
	mock.Mock
	Err error
}

func (m *MockClient) Integrations(_ string) ([]admin.Integration, error) {
	integration := admin.Integration{ID: "anId", Name: "aName", Provider: "google", Key: []byte("aKey")}
	return []admin.Integration{integration}, m.Err
}

func (m *MockClient) CreateIntegration(url string, _ string, _ []byte) error {
	args := m.Called(url)
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
	application := admin.Application{ID: "anId", IntegrationId: "anIntegrationId", ObjectId: "anObjectId", Name: "aName", Description: "aDescription"}
	return []admin.Application{application}, m.Err
}

func (m *MockClient) Application(_ string) (admin.Application, error) {
	return admin.Application{ID: "anId", IntegrationId: "anIntegrationId", ObjectId: "anObjectId", Name: "aName", Description: "aDescription"}, m.Err
}

func (m *MockClient) Policies(url string) ([]admin.Policy, string, error) {
	return []admin.Policy{
		{"aVersion", "anAction", admin.Subject{AuthenticatedUsers: []string{"aUser"}}, admin.Object{Resources: []string{"/"}}},
		{"aVersion", "anotherAction", admin.Subject{AuthenticatedUsers: []string{"anotherUser"}}, admin.Object{Resources: []string{"/"}}},
	}, "", m.Err
}

func (m *MockClient) Health(_ string) (string, error) {
	return "{\"status\":\"pass\"}", nil
}
