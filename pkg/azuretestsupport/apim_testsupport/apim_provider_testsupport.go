package apim_testsupport

import (
	"github.com/hexa-org/policy-orchestrator/internal/orchestratorproviders/microsoftazure"
	"github.com/hexa-org/policy-orchestrator/internal/orchestratorproviders/microsoftazure/azad"
	"github.com/hexa-org/policy-orchestrator/internal/orchestratorproviders/microsoftazure/azarm/azapim"
)

func NewAzureApimProvider(apimProviderService azapim.ArmApimSvc, azureClient azad.AzureClient) *microsoftazure.AzureApimProvider {
	provider := microsoftazure.NewAzureApimProvider(
		microsoftazure.WithArmApimSvcOverride(apimProviderService),
		microsoftazure.WithAzureClientOverride(azureClient))
	return provider
}
