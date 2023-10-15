package apimnv

import (
	"context"
	"github.com/Azure/azure-sdk-for-go/sdk/azcore"
	"github.com/Azure/azure-sdk-for-go/sdk/azcore/arm"
	"github.com/Azure/azure-sdk-for-go/sdk/azcore/runtime"
	azarmapim "github.com/Azure/azure-sdk-for-go/sdk/resourcemanager/apimanagement/armapimanagement"
)

type NamedValuesClient interface {
	NewListByServicePager(resourceGroupName, serviceName string, options *azarmapim.NamedValueClientListByServiceOptions) *runtime.Pager[azarmapim.NamedValueClientListByServiceResponse]
	BeginUpdate(ctx context.Context, resourceGroupName string, serviceName string, namedValueID string, ifMatch string, parameters azarmapim.NamedValueUpdateParameters, options *azarmapim.NamedValueClientBeginUpdateOptions) (*runtime.Poller[azarmapim.NamedValueClientUpdateResponse], error)
}

type namedValuesClient struct {
	internal *azarmapim.NamedValueClient
}

func NewNamedValuesClient(subscriptionID string, credential azcore.TokenCredential, options *arm.ClientOptions) NamedValuesClient {
	factory, _ := azarmapim.NewClientFactory(subscriptionID, credential, options)
	return &namedValuesClient{internal: factory.NewNamedValueClient()}
}

func (c *namedValuesClient) NewListByServicePager(resourceGroupName, serviceName string, options *azarmapim.NamedValueClientListByServiceOptions) *runtime.Pager[azarmapim.NamedValueClientListByServiceResponse] {
	return c.internal.NewListByServicePager(resourceGroupName, serviceName, options)
}

func (c *namedValuesClient) BeginUpdate(ctx context.Context, resourceGroup string, service string, nvName string, ifMatch string, updateParams azarmapim.NamedValueUpdateParameters, options *azarmapim.NamedValueClientBeginUpdateOptions) (*runtime.Poller[azarmapim.NamedValueClientUpdateResponse], error) {
	return c.internal.BeginUpdate(ctx, resourceGroup, service, nvName, ifMatch, updateParams, options)
}
