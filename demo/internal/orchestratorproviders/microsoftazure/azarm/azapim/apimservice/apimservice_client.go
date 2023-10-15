package apimservice

import (
	"github.com/Azure/azure-sdk-for-go/sdk/azcore"
	"github.com/Azure/azure-sdk-for-go/sdk/azcore/arm"
	"github.com/Azure/azure-sdk-for-go/sdk/azcore/runtime"
	azarmapim "github.com/Azure/azure-sdk-for-go/sdk/resourcemanager/apimanagement/armapimanagement"
)

type Client interface {
	NewListPager(options *azarmapim.ServiceClientListOptions) *runtime.Pager[azarmapim.ServiceClientListResponse]
}

type client struct {
	internal *azarmapim.ServiceClient
}

func NewClient(subscriptionID string, credential azcore.TokenCredential, options *arm.ClientOptions) Client {
	factory, _ := azarmapim.NewClientFactory(subscriptionID, credential, options)
	return &client{internal: factory.NewServiceClient()}
}

func (apiClient *client) NewListPager(options *azarmapim.ServiceClientListOptions) *runtime.Pager[azarmapim.ServiceClientListResponse] {
	return apiClient.internal.NewListPager(options)
}
