package apim_testsupport

import (
	"fmt"
	"github.com/hexa-org/policy-orchestrator/internal/orchestratorproviders/microsoftazure/azarm/armmodel"
	"github.com/hexa-org/policy-orchestrator/internal/orchestratorproviders/providerscommon"
	"github.com/stretchr/testify/mock"
	"strings"
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
	expReturnResourceRoles := make([]providerscommon.ResourceActionRoles, 0)

	for actionAndRes, roles := range retActionRoles {
		parts := strings.Split(actionAndRes, "/")
		resActionKey := fmt.Sprintf("resrol-http%s-%s", strings.ToLower(parts[0]), strings.Join(parts[1:], "-"))
		resRole := providerscommon.NewResourceActionRolesFromProviderValue(resActionKey, roles)
		expReturnResourceRoles = append(expReturnResourceRoles, resRole)
	}
	//resActionKey := "resrol-httpget-humanresources-us"
	//resRole := providerscommon.NewResourceActionRolesFromProviderValue(resActionKey, []string{"some-role"})
	m.On("GetResourceRoles", serviceInfo).
		Return(expReturnResourceRoles, nil)
}
