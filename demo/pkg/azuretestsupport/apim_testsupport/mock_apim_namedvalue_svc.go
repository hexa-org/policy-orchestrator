package apim_testsupport

import (
	"github.com/hexa-org/policy-orchestrator/demo/internal/orchestratorproviders/microsoftazure/azarm/armmodel"
	"github.com/hexa-org/policy-orchestrator/demo/internal/orchestratorproviders/providerscommon"
	"github.com/hexa-org/policy-orchestrator/demo/pkg/testsupport/policytestsupport"
	"github.com/stretchr/testify/mock"
)

type MockApimNamedValueSvc struct {
	mock.Mock
}

func NewMockApimNamedValueSvc() *MockApimNamedValueSvc {
	return &MockApimNamedValueSvc{}
}

func (m *MockApimNamedValueSvc) GetResourceRoles(s armmodel.ApimServiceInfo) ([]providerscommon.ResourceActionRoles, error) {
	returnArgs := m.Called(s)
	return returnArgs.Get(0).([]providerscommon.ResourceActionRoles), returnArgs.Error(1)
}

func (m *MockApimNamedValueSvc) UpdateResourceRole(s armmodel.ApimServiceInfo, nv providerscommon.ResourceActionRoles) error {
	returnArgs := m.Called(s, nv)
	return returnArgs.Error(0)
}

func (m *MockApimNamedValueSvc) ExpectGetResourceRoles(serviceInfo armmodel.ApimServiceInfo, retActionRoles map[string][]string) {
	expReturnResourceRoles := policytestsupport.MakeRarList(retActionRoles)
	m.On("GetResourceRoles", serviceInfo).
		Return(expReturnResourceRoles, nil)
}
