package microsoftazure_test

import (
	"github.com/hexa-org/policy-orchestrator/internal/orchestrator"
	"github.com/hexa-org/policy-orchestrator/internal/orchestratorproviders/microsoftazure"
	"github.com/hexa-org/policy-orchestrator/internal/orchestratorproviders/microsoftazure/azarm/armmodel"
	"github.com/hexa-org/policy-orchestrator/pkg/azuretestsupport"
	"github.com/hexa-org/policy-orchestrator/pkg/azuretestsupport/apim_testsupport"
	"github.com/hexa-org/policy-orchestrator/pkg/azuretestsupport/armtestsupport"
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestNewAzureApimProvider(t *testing.T) {
	provider := microsoftazure.NewAzureApimProvider()
	assert.NotNil(t, provider)
}

func TestDiscoverApplications_Success(t *testing.T) {
	apimSvc := apim_testsupport.NewMockArmApimSvc()
	azureClient := azuretestsupport.NewMockAzureClient()
	provider := apim_testsupport.NewAzureApimProvider(apimSvc, azureClient)

	key := azuretestsupport.AzureKeyBytes()
	info := orchestrator.IntegrationInfo{Name: "azure", Key: key}

	azureClient.ExpectGetAzureApplications()
	apimSvc.ExpectGetApimServiceInfo(armtestsupport.ApimServiceGatewayUrl)

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
	provider := apim_testsupport.NewAzureApimProvider(apimSvc, azureClient)

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
