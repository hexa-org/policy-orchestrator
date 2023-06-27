package apimnv

import (
	"context"
	"github.com/Azure/azure-sdk-for-go/sdk/azcore"
	"github.com/Azure/azure-sdk-for-go/sdk/azcore/arm"
	"github.com/Azure/azure-sdk-for-go/sdk/azcore/runtime"
	apim "github.com/Azure/azure-sdk-for-go/sdk/resourcemanager/apimanagement/armapimanagement"
)

type NamedValuesClient interface {
	NewListByServicePager(resourceGroupName, serviceName string, options *apim.NamedValueClientListByServiceOptions) *runtime.Pager[apim.NamedValueClientListByServiceResponse]
	BeginUpdate(ctx context.Context, resourceGroupName string, serviceName string, namedValueID string, ifMatch string, parameters apim.NamedValueUpdateParameters, options *apim.NamedValueClientBeginUpdateOptions) (*runtime.Poller[apim.NamedValueClientUpdateResponse], error)
}

type namedValuesClient struct {
	internal *apim.NamedValueClient
}

func NewNamedValuesClient(subscriptionID string, credential azcore.TokenCredential, options *arm.ClientOptions) NamedValuesClient {
	factory, _ := apim.NewClientFactory(subscriptionID, credential, options)
	return &namedValuesClient{internal: factory.NewNamedValueClient()}
}

func (c *namedValuesClient) NewListByServicePager(resourceGroupName, serviceName string, options *apim.NamedValueClientListByServiceOptions) *runtime.Pager[apim.NamedValueClientListByServiceResponse] {
	return c.internal.NewListByServicePager(resourceGroupName, serviceName, options)
}

func (c *namedValuesClient) BeginUpdate(ctx context.Context, resourceGroup string, service string, nvName string, ifMatch string, updateParams apim.NamedValueUpdateParameters, options *apim.NamedValueClientBeginUpdateOptions) (*runtime.Poller[apim.NamedValueClientUpdateResponse], error) {
	return c.internal.BeginUpdate(ctx, resourceGroup, service, nvName, ifMatch, updateParams, options)
}
