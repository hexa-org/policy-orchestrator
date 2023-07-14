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
	m.On("GetApimServiceInfo", serviceInfo.ServiceUrl).
		Return(serviceInfo, nil)
}
