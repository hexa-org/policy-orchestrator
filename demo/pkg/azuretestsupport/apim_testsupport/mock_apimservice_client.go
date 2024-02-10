package apim_testsupport

import (
	"github.com/Azure/azure-sdk-for-go/sdk/azcore/runtime"
	azarmapim "github.com/Azure/azure-sdk-for-go/sdk/resourcemanager/apimanagement/armapimanagement"
	"github.com/hexa-org/policy-orchestrator/demo/pkg/azuretestsupport/armtestsupport"
	"github.com/stretchr/testify/mock"
)

type MockApimServiceClient struct {
	mock.Mock
}

func NewMockApimServiceClient() *MockApimServiceClient {
	return &MockApimServiceClient{}
}

func (m *MockApimServiceClient) NewListPager(options *azarmapim.ServiceClientListOptions) *runtime.Pager[azarmapim.ServiceClientListResponse] {
	returnArgs := m.Called(options)
	return returnArgs.Get(0).(*runtime.Pager[azarmapim.ServiceClientListResponse])
}

func (m *MockApimServiceClient) ExpectNewListPager(pages ...azarmapim.ServiceClientListResponse) {
	numPages := len(pages)
	pagerBuilder := armtestsupport.NewFakeCountBasedPagerBuilder[azarmapim.ServiceClientListResponse](numPages)

	for _, pg := range pages {
		pagerBuilder.AddPage(pg)
	}

	var options *azarmapim.ServiceClientListOptions
	m.On("NewListPager", options).
		Return(pagerBuilder.Pager())
}
