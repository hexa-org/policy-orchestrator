package admin_test

import (
	"github.com/hexa-org/policy-orchestrator/pkg/admin"
	"github.com/stretchr/testify/mock"
)

type MockClient struct {
	mock.Mock
	Err error
}

func (m *MockClient) Integrations(url string) ([]admin.Integration, error) {
	return []admin.Integration{{"anId", "aName", "google", []byte("aKey")}}, m.Err
}

func (m *MockClient) CreateIntegration(url string, provider string, key []byte) error {
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

func (m *MockClient) Applications(url string) ([]admin.Application, error) {
	return []admin.Application{{"anId", "anIntegrationId", "anObjectId", "aName", "aDescription"}}, m.Err
}

func (m *MockClient) Application(url string) (admin.Application, error) {
	return admin.Application{ID: "anId", IntegrationId: "anIntegrationId", ObjectId: "anObjectId", Name: "aName", Description: "aDescription"}, m.Err
}

func (m *MockClient) Health(url string) (string, error) {
	return "{\"status\":\"pass\"}", nil
}
