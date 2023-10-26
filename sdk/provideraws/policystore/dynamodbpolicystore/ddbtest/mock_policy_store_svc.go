package ddbtest

import (
	"github.com/hexa-org/policy-orchestrator/sdk/core/idp"
	"github.com/hexa-org/policy-orchestrator/sdk/core/rar"
	"github.com/stretchr/testify/mock"
)

type MockPolicyStoreSvc[R rar.ResourceActionRolesMapper] struct {
	mock.Mock
}

func (m *MockPolicyStoreSvc[R]) GetPolicies(app idp.AppInfo) ([]rar.ResourceActionRoles, error) {
	args := m.Called(app)
	return args.Get(0).([]rar.ResourceActionRoles), args.Error(1)
}

func (m *MockPolicyStoreSvc[R]) SetPolicy(rar rar.ResourceActionRoles) error {
	args := m.Called(rar)
	return args.Error(0)
}
