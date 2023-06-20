package azureapim

import (
	"github.com/Azure/azure-sdk-for-go/sdk/azcore"
	"github.com/Azure/azure-sdk-for-go/sdk/azcore/arm"
	"github.com/Azure/azure-sdk-for-go/sdk/azcore/runtime"
	"github.com/Azure/azure-sdk-for-go/sdk/resourcemanager/apimanagement/armapimanagement"
	log "golang.org/x/exp/slog"
)

type ArmApimApiClient interface {
	NewListByServicePager(resourceGroupName string, serviceName string, options *armapimanagement.APIClientListByServiceOptions) *runtime.Pager[armapimanagement.APIClientListByServiceResponse]
	//Get(ctx context.Context, resourceGroupName string, serviceName string, apiID string, options *armapimanagement.APIClientGetOptions) (armapimanagement.APIClientGetResponse, error)
}

type armApimApiClient struct {
	internal *armapimanagement.APIClient
}

func newApimApiClient(subscriptionID string, credential azcore.TokenCredential, options *arm.ClientOptions) (ArmApimApiClient, error) {
	factory, err := armapimanagement.NewClientFactory(subscriptionID, credential, options)
	// TODO - above call doesnt seem to be throwing any errors. See TestNewArmApimSvc
	if err != nil {
		log.Error("Error from armapimanagement.NewClientFactory. Error=", err)
		return nil, err
	}

	return &armApimApiClient{internal: factory.NewAPIClient()}, nil
}

/*
func (apiClient *armApimApiClient) Get(ctx context.Context, resourceGroupName string, serviceName string, apiID string, options *armapimanagement.APIClientGetOptions) (armapimanagement.APIClientGetResponse, error) {
	return apiClient.internal.Get(ctx, resourceGroupName, serviceName, apiID, options)
}
*/

func (apiClient *armApimApiClient) NewListByServicePager(resourceGroupName string, serviceName string, options *armapimanagement.APIClientListByServiceOptions) *runtime.Pager[armapimanagement.APIClientListByServiceResponse] {
	return apiClient.internal.NewListByServicePager(resourceGroupName, serviceName, options)
}
