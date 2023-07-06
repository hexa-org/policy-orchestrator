package armtestsupport

import (
	"context"
	"github.com/Azure/azure-sdk-for-go/sdk/azcore/runtime"
	azarmapim "github.com/Azure/azure-sdk-for-go/sdk/resourcemanager/apimanagement/armapimanagement"
	"github.com/stretchr/testify/mock"
)

type MockNamedValuesClient struct {
	mock.Mock
}

func NewMockNamedValuesClient() *MockNamedValuesClient {
	return &MockNamedValuesClient{}
}

func (m *MockNamedValuesClient) NewListByServicePager(resourceGroupName, serviceName string, options *azarmapim.NamedValueClientListByServiceOptions) *runtime.Pager[azarmapim.NamedValueClientListByServiceResponse] {
	returnArgs := m.Called(resourceGroupName, serviceName, options)
	return returnArgs.Get(0).(*runtime.Pager[azarmapim.NamedValueClientListByServiceResponse])
}

func (m *MockNamedValuesClient) BeginUpdate(ctx context.Context, resourceGroupName string, serviceName string, namedValueID string, ifMatch string, parameters azarmapim.NamedValueUpdateParameters, _ *azarmapim.NamedValueClientBeginUpdateOptions) (*runtime.Poller[azarmapim.NamedValueClientUpdateResponse], error) {
	returnArgs := m.Called(ctx, resourceGroupName, serviceName, namedValueID, ifMatch, parameters, nil)
	return returnArgs.Get(0).(*runtime.Poller[azarmapim.NamedValueClientUpdateResponse]), returnArgs.Error(1)
}
