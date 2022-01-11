package admin_test

import (
	"github.com/stretchr/testify/mock"
	"hexa/pkg/admin"
)

type MockClient struct {
	mock.Mock
}

func (m *MockClient) Integrations(url string) ([]admin.Integration, error) {
	return []admin.Integration{{"anId", "aName", "google", []byte("aKey")}}, nil
}

func (m *MockClient) CreateIntegration(url string, provider string, key []byte) error {
	args := m.Called(url)
	if len(args) > 0 {
		return args.Error(0)
	}
	return nil
}

func (m *MockClient) DeleteIntegration(url string) error {
	args := m.Called(url)
	if len(args) > 0 {
		return args.Error(0)
	}
	return nil
}

func (m *MockClient) Applications(url string) ([]admin.Application, error) {
	return []admin.Application{{"anId", "anIntegrationId", "anObjectId", "aName", "aDescription"}}, nil
}

func (m *MockClient) Health(url string) (string, error) {
	return "{\"status\":\"pass\"}", nil
}
