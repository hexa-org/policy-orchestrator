package azureresource

import (
	"github.com/Azure/azure-sdk-for-go/sdk/azcore"
	"github.com/Azure/azure-sdk-for-go/sdk/azcore/arm"
	"github.com/Azure/azure-sdk-for-go/sdk/azcore/runtime"
	"github.com/Azure/azure-sdk-for-go/sdk/resourcemanager/resources/armresources"
	log "golang.org/x/exp/slog"
)

type ArmResourcesClient interface {
	NewListPager(options *armresources.ClientListOptions) *runtime.Pager[armresources.ClientListResponse]
	//GetApiManagementResources() (map[string]ArmResource, error)
}

type armResourcesClient struct {
	internal *armresources.Client
}

func NewArmResourcesClient(internal *armresources.Client) ArmResourcesClient {
	return &armResourcesClient{internal: internal}
}

func newArmResourcesClient(subscriptionID string, credential azcore.TokenCredential, options *arm.ClientOptions) (ArmResourcesClient, error) {
	factory, err := armresources.NewClientFactory(subscriptionID, credential, options)
	if err != nil {
		log.Error("Error arm resources.NewClientFactory.", err)
		return nil, err
	}

	return NewArmResourcesClient(factory.NewClient()), nil
}

func (c *armResourcesClient) NewListPager(options *armresources.ClientListOptions) *runtime.Pager[armresources.ClientListResponse] {
	return c.internal.NewListPager(options)
}
