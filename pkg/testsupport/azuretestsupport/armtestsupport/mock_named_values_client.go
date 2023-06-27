package armtestsupport

import (
	"context"
	"github.com/Azure/azure-sdk-for-go/sdk/azcore/runtime"
	apim "github.com/Azure/azure-sdk-for-go/sdk/resourcemanager/apimanagement/armapimanagement"
	"github.com/stretchr/testify/mock"
)

type MockNamedValuesClient struct {
	mock.Mock
}

func NewMockNamedValuesClient() *MockNamedValuesClient {
	return &MockNamedValuesClient{}
}

func (m *MockNamedValuesClient) NewListByServicePager(resourceGroupName, serviceName string, options *apim.NamedValueClientListByServiceOptions) *runtime.Pager[apim.NamedValueClientListByServiceResponse] {
	returnArgs := m.Called(resourceGroupName, serviceName, options)
	return returnArgs.Get(0).(*runtime.Pager[apim.NamedValueClientListByServiceResponse])
}

func (m *MockNamedValuesClient) BeginUpdate(ctx context.Context, resourceGroupName string, serviceName string, namedValueID string, ifMatch string, parameters apim.NamedValueUpdateParameters, _ *apim.NamedValueClientBeginUpdateOptions) (*runtime.Poller[apim.NamedValueClientUpdateResponse], error) {
	returnArgs := m.Called(ctx, resourceGroupName, serviceName, namedValueID, ifMatch, parameters, nil)
	return returnArgs.Get(0).(*runtime.Poller[apim.NamedValueClientUpdateResponse]), returnArgs.Error(1)
}
