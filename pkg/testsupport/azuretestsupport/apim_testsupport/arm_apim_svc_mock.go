package apim_testsupport

import (
	"github.com/hexa-org/policy-orchestrator/internal/orchestratorproviders/microsoftazure"
	"github.com/hexa-org/policy-orchestrator/internal/orchestratorproviders/microsoftazure/armmodel"
	"github.com/hexa-org/policy-orchestrator/internal/orchestratorproviders/microsoftazure/azureapim"
	"github.com/hexa-org/policy-orchestrator/pkg/testsupport"
	"github.com/hexa-org/policy-orchestrator/pkg/testsupport/azuretestsupport"
	"github.com/hexa-org/policy-orchestrator/pkg/testsupport/azuretestsupport/armtestsupport"
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

func (m *MockArmApimSvc) GetResourceRoles(s armmodel.ApimServiceInfo) ([]armmodel.ResourceActionRoles, error) {
	returnArgs := m.Called(s)
	return returnArgs.Get(0).([]armmodel.ResourceActionRoles), returnArgs.Error(1)
}

func (m *MockArmApimSvc) ExpectGetApimServiceInfo(gatewayUrl string) {
	expServiceInfo := armmodel.ApimServiceInfo{
		ArmResource: armmodel.ArmResource{
			FullyQualifiedId: ApimServiceId(),
			ResourceGroup:    armtestsupport.ApimResourceGroupName,
			Type:             "Microsoft.ApiManagement/service",
			Name:             armtestsupport.ApimServiceName,
			DisplayName:      armtestsupport.ApimServiceName,
		},
		ServiceUrl: armtestsupport.ApimServiceGatewayUrl,
	}

	m.On("GetApimServiceInfo", gatewayUrl).
		Return(expServiceInfo, nil)
}

//func (m *MockArmApimSvc) getApimApiInfo(armResource armmodel.ArmResource, serviceUrl string) (*azureapim.ApimServiceInfo, error) {
//	returnArgs := m.Called(armResource, serviceUrl)
//	return returnArgs.Get(0).(*azureapim.ApimServiceInfo), returnArgs.Error(1)
//}

func BuildApimSvc(mockHttpClient *testsupport.MockHTTPClient) azureapim.ArmApimSvc {
	key := azuretestsupport.AzureClientKey()
	factory, _ := microsoftazure.NewApimProviderSvcFactory(key, mockHttpClient)
	service, _ := factory.NewApimSvc()
	return service
}
