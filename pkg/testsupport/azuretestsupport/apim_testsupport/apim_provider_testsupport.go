package apim_testsupport

import (
	"github.com/hexa-org/policy-orchestrator/internal/orchestratorproviders/microsoftazure"
	"github.com/hexa-org/policy-orchestrator/internal/orchestratorproviders/microsoftazure/azureapim"
)

func NewAzureApimProvider(apimProviderService azureapim.ArmApimSvc, azureClient microsoftazure.AzureClient) *microsoftazure.AzureApimProvider {
	provider := microsoftazure.NewAzureApimProvider(
		microsoftazure.WithArmApimSvcOverride(apimProviderService),
		microsoftazure.WithAzureClientOverride(azureClient))
	return provider
}
