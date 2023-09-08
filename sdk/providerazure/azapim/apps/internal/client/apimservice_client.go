package client

import (
	"context"
	"fmt"
	"github.com/Azure/azure-sdk-for-go/sdk/azcore/runtime"
	azarmapim "github.com/Azure/azure-sdk-for-go/sdk/resourcemanager/apimanagement/armapimanagement"
	"github.com/hexa-org/policy-orchestrator/sdk/providerazure/azapim/internal/clientsupport"
	"github.com/hexa-org/policy-orchestrator/sdk/providerazure/azurecommon"
)

type ApimServiceClient interface {
	NewListPager(options *azarmapim.ServiceClientListOptions) *runtime.Pager[azarmapim.ServiceClientListResponse]
	Get(ctx context.Context, resourceGroupName string, serviceName string, options *azarmapim.ServiceClientGetOptions) (azarmapim.ServiceClientGetResponse, error)
}

type client struct {
	internal *azarmapim.ServiceClient
}

func NewApimServiceClient(key []byte, httpClient azurecommon.HTTPClient) (ApimServiceClient, error) {
	azKey, err := azurecommon.DecodeKey(key)
	if err != nil {
		return nil, fmt.Errorf("failed to create azure ApimServiceClient. Unable to decode AzureKey. error: %w", err)
	}

	creds, err := azurecommon.ClientSecretCredentialsWithAzureKey(azKey, httpClient)

	if err != nil {
		return nil, fmt.Errorf("failed to create azure ApimServiceClient. error: %w", err)
	}

	armClientOptions := clientsupport.NewArmClientOptions(httpClient)

	factory, _ := azarmapim.NewClientFactory(azKey.Subscription, creds, armClientOptions)
	return &client{internal: factory.NewServiceClient()}, nil
}

func (apiClient *client) NewListPager(options *azarmapim.ServiceClientListOptions) *runtime.Pager[azarmapim.ServiceClientListResponse] {
	return apiClient.internal.NewListPager(options)
}

func (apiClient *client) Get(ctx context.Context, resourceGroupName string, serviceName string, options *azarmapim.ServiceClientGetOptions) (azarmapim.ServiceClientGetResponse, error) {
	return apiClient.internal.Get(ctx, resourceGroupName, serviceName, options)
}
