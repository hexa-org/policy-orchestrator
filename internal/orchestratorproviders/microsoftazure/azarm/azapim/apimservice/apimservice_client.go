package apimservice

import (
	"github.com/Azure/azure-sdk-for-go/sdk/azcore"
	"github.com/Azure/azure-sdk-for-go/sdk/azcore/arm"
	"github.com/Azure/azure-sdk-for-go/sdk/azcore/runtime"
	"github.com/Azure/azure-sdk-for-go/sdk/resourcemanager/apimanagement/armapimanagement"
)

type Client interface {
	List(options *armapimanagement.ServiceClientListOptions) *runtime.Pager[armapimanagement.ServiceClientListResponse]
}

type client struct {
	internal *armapimanagement.ServiceClient
}

func NewClient(subscriptionID string, credential azcore.TokenCredential, options *arm.ClientOptions) (Client, error) {
	factory, _ := armapimanagement.NewClientFactory(subscriptionID, credential, options)
	// TODO - probaby remove the error return, this func doesnt thrown any errors
	// See TestNewArmApimSvc
	return &client{internal: factory.NewServiceClient()}, nil
}

func (apiClient *client) List(options *armapimanagement.ServiceClientListOptions) *runtime.Pager[armapimanagement.ServiceClientListResponse] {
	return apiClient.internal.NewListPager(options)
}
