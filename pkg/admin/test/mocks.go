package admin_test

import (
	"github.com/stretchr/testify/mock"
	"hexa/pkg/admin"
)

type MockClient struct {
	mock.Mock
}

func (m *MockClient) Applications(url string) ([]admin.Application, error) {
	return []admin.Application{{"anApp"}}, nil
}

func (m *MockClient) Health(url string) (string, error) {
	return "{\"status\":\"pass\"}", nil
}
