package clientsupport

import (
	"github.com/Azure/azure-sdk-for-go/sdk/azcore/arm"
	"github.com/Azure/azure-sdk-for-go/sdk/azcore/policy"
	"github.com/hexa-org/policy-orchestrator/sdk/providerazure/azurecommon"
)

func NewArmClientOptions(httpClient azurecommon.HTTPClient) *arm.ClientOptions {
	var clientOpts *arm.ClientOptions
	if httpClient != nil {
		clientOpts = &arm.ClientOptions{
			ClientOptions: policy.ClientOptions{
				Retry:     policy.RetryOptions{MaxRetries: -1},
				Transport: httpClient,
			},
		}
	}
	return clientOpts
}
