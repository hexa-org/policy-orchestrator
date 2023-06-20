package apimnv

import (
	"github.com/Azure/azure-sdk-for-go/sdk/azcore"
	"github.com/Azure/azure-sdk-for-go/sdk/azcore/arm"
	"github.com/Azure/azure-sdk-for-go/sdk/azcore/runtime"
	apim "github.com/Azure/azure-sdk-for-go/sdk/resourcemanager/apimanagement/armapimanagement"
)

type NamedValuesClient interface {
	List(resourceGroupName, serviceName string, options *apim.NamedValueClientListByServiceOptions) *runtime.Pager[apim.NamedValueClientListByServiceResponse]
}

type namedValuesClient struct {
	internal *apim.NamedValueClient
}

func NewNamedValuesClient(subscriptionID string, credential azcore.TokenCredential, options *arm.ClientOptions) NamedValuesClient {
	factory, _ := apim.NewClientFactory(subscriptionID, credential, options)
	return &namedValuesClient{internal: factory.NewNamedValueClient()}
}

func (c *namedValuesClient) List(resourceGroupName, serviceName string, options *apim.NamedValueClientListByServiceOptions) *runtime.Pager[apim.NamedValueClientListByServiceResponse] {
	return c.internal.NewListByServicePager(resourceGroupName, serviceName, options)
}
