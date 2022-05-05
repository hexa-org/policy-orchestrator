package decisionproviders

import (
	"github.com/stretchr/testify/mock"
	"net/http"
)

type MockDecisionProvider struct {
	mock.Mock
	BuildErr error
	Decision bool
	AllowErr error
}

func (m *MockDecisionProvider) Allow(_ interface{}) (bool, error) {
	m.Called()
	return m.Decision, m.AllowErr
}

func (m *MockDecisionProvider) BuildInput(r *http.Request) (interface{}, error) {
	m.Called()

	return map[string]string{
		"method": "GET",
		"path":   r.URL.Path,
	}, m.BuildErr
}
