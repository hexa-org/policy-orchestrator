package azureapim

import (
	"github.com/Azure/azure-sdk-for-go/sdk/azcore"
	"github.com/Azure/azure-sdk-for-go/sdk/azcore/arm"
	"github.com/Azure/azure-sdk-for-go/sdk/azcore/runtime"
	"github.com/Azure/azure-sdk-for-go/sdk/resourcemanager/apimanagement/armapimanagement"
)

type ApimServiceClient interface {
	List(options *armapimanagement.ServiceClientListOptions) *runtime.Pager[armapimanagement.ServiceClientListResponse]
}

type armServiceClient struct {
	internal *armapimanagement.ServiceClient
}

func newArmServiceClient(subscriptionID string, credential azcore.TokenCredential, options *arm.ClientOptions) (ApimServiceClient, error) {
	factory, _ := armapimanagement.NewClientFactory(subscriptionID, credential, options)
	// TODO - probaby remove the error return, this func doesnt thrown any errors
	// See TestNewArmApimSvc
	return &armServiceClient{internal: factory.NewServiceClient()}, nil
}

func (apiClient *armServiceClient) List(options *armapimanagement.ServiceClientListOptions) *runtime.Pager[armapimanagement.ServiceClientListResponse] {
	return apiClient.internal.NewListPager(options)
}
