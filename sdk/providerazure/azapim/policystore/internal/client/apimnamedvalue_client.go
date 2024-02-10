package client

import (
	"context"
	"fmt"
	"github.com/Azure/azure-sdk-for-go/sdk/azcore"
	"github.com/Azure/azure-sdk-for-go/sdk/azcore/arm"
	"github.com/Azure/azure-sdk-for-go/sdk/azcore/runtime"
	azarmapim "github.com/Azure/azure-sdk-for-go/sdk/resourcemanager/apimanagement/armapimanagement"
	"github.com/hexa-org/policy-orchestrator/sdk/providerazure/azapim/internal/clientsupport"
	"github.com/hexa-org/policy-orchestrator/sdk/providerazure/azurecommon"
)

type NamedValuesClient interface {
	NewListByServicePager(resourceGroupName, serviceName string, options *azarmapim.NamedValueClientListByServiceOptions) *runtime.Pager[azarmapim.NamedValueClientListByServiceResponse]
	BeginUpdate(ctx context.Context, resourceGroupName string, serviceName string, namedValueID string, ifMatch string, parameters azarmapim.NamedValueUpdateParameters, options *azarmapim.NamedValueClientBeginUpdateOptions) (*runtime.Poller[azarmapim.NamedValueClientUpdateResponse], error)
}

type namedValuesClient struct {
	internal *azarmapim.NamedValueClient
}

func NewNamedValuesClient(key []byte, httpClient azurecommon.HTTPClient) (NamedValuesClient, error) {
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
	return &namedValuesClient{internal: factory.NewNamedValueClient()}, nil
}

func NewNamedValuesClientOld(subscriptionID string, credential azcore.TokenCredential, options *arm.ClientOptions) NamedValuesClient {
	factory, _ := azarmapim.NewClientFactory(subscriptionID, credential, options)
	return &namedValuesClient{internal: factory.NewNamedValueClient()}
}

func (c *namedValuesClient) NewListByServicePager(resourceGroupName, serviceName string, options *azarmapim.NamedValueClientListByServiceOptions) *runtime.Pager[azarmapim.NamedValueClientListByServiceResponse] {
	return c.internal.NewListByServicePager(resourceGroupName, serviceName, options)
}

func (c *namedValuesClient) BeginUpdate(ctx context.Context, resourceGroup string, service string, nvName string, ifMatch string, updateParams azarmapim.NamedValueUpdateParameters, options *azarmapim.NamedValueClientBeginUpdateOptions) (*runtime.Poller[azarmapim.NamedValueClientUpdateResponse], error) {
	return c.internal.BeginUpdate(ctx, resourceGroup, service, nvName, ifMatch, updateParams, options)
}
