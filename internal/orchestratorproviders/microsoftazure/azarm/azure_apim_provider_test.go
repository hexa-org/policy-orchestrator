package azarm_test

import (
	"github.com/hexa-org/policy-orchestrator/internal/orchestrator"
	"github.com/hexa-org/policy-orchestrator/internal/orchestratorproviders/microsoftazure/azad"
	"github.com/hexa-org/policy-orchestrator/internal/orchestratorproviders/microsoftazure/azarm"
	"github.com/hexa-org/policy-orchestrator/internal/orchestratorproviders/microsoftazure/azarm/armmodel"
	"github.com/hexa-org/policy-orchestrator/internal/orchestratorproviders/microsoftazure/azarm/azapim"
	"github.com/hexa-org/policy-orchestrator/internal/orchestratorproviders/microsoftazure/azarm/azapim/apimnv"
	"github.com/hexa-org/policy-orchestrator/pkg/azuretestsupport"
	"github.com/hexa-org/policy-orchestrator/pkg/azuretestsupport/apim_testsupport"
	"github.com/hexa-org/policy-orchestrator/pkg/azuretestsupport/armtestsupport"
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestNewAzureApimProvider(t *testing.T) {
	provider := azarm.NewAzureApimProvider()
	assert.NotNil(t, provider)
}

func newAzureApimProvider(apimProviderService azapim.ArmApimSvc, azureClient azad.AzureClient, apimNvSvc apimnv.ApimNamedValueSvc) *azarm.AzureApimProvider {
	provider := azarm.NewAzureApimProvider(
		azarm.WithArmApimSvcOverride(apimProviderService),
		azarm.WithAzureClientOverride(azureClient),
		azarm.WithApimNamedValueSvcOverride(apimNvSvc))
	return provider
}

func TestDiscoverApplications_Success(t *testing.T) {
	apimSvc := apim_testsupport.NewMockArmApimSvc()
	azureClient := azuretestsupport.NewMockAzureClient()
	apimNvSvc := apim_testsupport.NewMockApimNamedValueSvc()
	provider := newAzureApimProvider(apimSvc, azureClient, apimNvSvc)

	key := azuretestsupport.AzureKeyBytes()
	info := orchestrator.IntegrationInfo{Name: "azure", Key: key}

	azureClient.ExpectGetAzureApplications()
	serviceInfo := apim_testsupport.ApimServiceInfo(armtestsupport.ApimServiceGatewayUrl)
	apimSvc.ExpectGetApimServiceInfo(serviceInfo)

	applications, err := provider.DiscoverApplications(info)
	assert.NoError(t, err)
	assert.NotNil(t, applications)
	assert.Equal(t, 1, len(applications))

	expApp := orchestrator.ApplicationInfo{
		ObjectID:    azuretestsupport.ServicePrincipalId,
		Name:        azuretestsupport.AzureAppName,
		Description: azuretestsupport.AzureAppId,
		Service:     armtestsupport.ApimServiceGatewayUrl,
	}

	assert.Equal(t, expApp, applications[0])
}

func TestDiscoverApplications_NoApimServices(t *testing.T) {
	apimSvc := apim_testsupport.NewMockArmApimSvc()
	azureClient := azuretestsupport.NewMockAzureClient()
	apimNvSvc := apim_testsupport.NewMockApimNamedValueSvc()
	provider := newAzureApimProvider(apimSvc, azureClient, apimNvSvc)

	key := azuretestsupport.AzureKeyBytes()
	info := orchestrator.IntegrationInfo{Name: "azure", Key: key}

	azureClient.ExpectGetAzureApplications()
	apimSvc.On("GetApimServiceInfo", armtestsupport.ApimServiceGatewayUrl).
		Return(armmodel.ApimServiceInfo{}, nil)

	applications, err := provider.DiscoverApplications(info)
	assert.NoError(t, err)
	assert.NotNil(t, applications)
	assert.Equal(t, []orchestrator.ApplicationInfo{}, applications)
}
