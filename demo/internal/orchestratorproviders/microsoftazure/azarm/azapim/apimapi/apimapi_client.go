package apimapi

import (
	"github.com/Azure/azure-sdk-for-go/sdk/azcore"
	"github.com/Azure/azure-sdk-for-go/sdk/azcore/arm"
	"github.com/Azure/azure-sdk-for-go/sdk/azcore/runtime"
	azarmapim "github.com/Azure/azure-sdk-for-go/sdk/resourcemanager/apimanagement/armapimanagement"
)

type ArmApimApiClient interface {
	NewListByServicePager(resourceGroupName string, serviceName string, options *azarmapim.APIClientListByServiceOptions) *runtime.Pager[azarmapim.APIClientListByServiceResponse]
	//Get(ctx context.Context, resourceGroupName string, serviceName string, apiID string, options *armapimanagement.APIClientGetOptions) (armapimanagement.APIClientGetResponse, error)
}

type armApimApiClient struct {
	internal *azarmapim.APIClient
}

func NewApimApiClient(subscriptionID string, credential azcore.TokenCredential, options *arm.ClientOptions) ArmApimApiClient {
	factory, _ := azarmapim.NewClientFactory(subscriptionID, credential, options)
	return &armApimApiClient{internal: factory.NewAPIClient()}
}

/*
func (apiClient *armApimApiClient) Get(ctx context.Context, resourceGroupName string, serviceName string, apiID string, options *armapimanagement.APIClientGetOptions) (armapimanagement.APIClientGetResponse, error) {
	return apiClient.internal.Get(ctx, resourceGroupName, serviceName, apiID, options)
}
*/

func (apiClient *armApimApiClient) NewListByServicePager(resourceGroupName string, serviceName string, options *azarmapim.APIClientListByServiceOptions) *runtime.Pager[azarmapim.APIClientListByServiceResponse] {
	return apiClient.internal.NewListByServicePager(resourceGroupName, serviceName, options)
}
