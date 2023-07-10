package apim_testsupport

import (
	"github.com/hexa-org/policy-orchestrator/internal/orchestratorproviders/microsoftazure/azarm/armmodel"
	"github.com/stretchr/testify/mock"
)

type MockArmApimSvc struct {
	mock.Mock
}

func NewMockArmApimSvc() *MockArmApimSvc {
	return &MockArmApimSvc{}
}

func (m *MockArmApimSvc) GetApimServiceInfo(serviceUrl string) (armmodel.ApimServiceInfo, error) {
	returnArgs := m.Called(serviceUrl)
	return returnArgs.Get(0).(armmodel.ApimServiceInfo), returnArgs.Error(1)
}

func (m *MockArmApimSvc) ExpectGetApimServiceInfo(serviceInfo armmodel.ApimServiceInfo) {
	//expServiceInfo := armmodel.ApimServiceInfo{
	//	ArmResource: armmodel.ArmResource{
	//		FullyQualifiedId: ApimServiceId(),
	//		ResourceGroup:    armtestsupport.ApimResourceGroupName,
	//		Type:             "Microsoft.ApiManagement/service",
	//		Name:             armtestsupport.ApimServiceName,
	//		DisplayName:      armtestsupport.ApimServiceName,
	//	},
	//	ServiceUrl: armtestsupport.ApimServiceGatewayUrl,
	//}

	m.On("GetApimServiceInfo", serviceInfo.ServiceUrl).
		Return(serviceInfo, nil)
}

//func (m *MockArmApimSvc) getApimApiInfo(armResource armmodel.ArmResource, serviceUrl string) (*azureapim.ApimServiceInfo, error) {
//	returnArgs := m.Called(armResource, serviceUrl)
//	return returnArgs.Get(0).(*azureapim.ApimServiceInfo), returnArgs.Error(1)
//}

/*

func (m *MockArmApimSvc) GetResourceRoles(s armmodel.ApimServiceInfo) ([]providerscommon.ResourceActionRoles, error) {
	returnArgs := m.Called(s)
	return returnArgs.Get(0).([]providerscommon.ResourceActionRoles), returnArgs.Error(1)
}

func (m *MockArmApimSvc) UpdateResourceRole(s armmodel.ApimServiceInfo, nv providerscommon.ResourceActionRoles) error {
	returnArgs := m.Called(s, nv)
	return returnArgs.Error(0)
}
*/
