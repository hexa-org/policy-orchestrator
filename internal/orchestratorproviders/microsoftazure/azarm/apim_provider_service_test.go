package azarm_test

import (
	"github.com/hexa-org/policy-orchestrator/internal/orchestrator"
	"github.com/hexa-org/policy-orchestrator/internal/orchestratorproviders/microsoftazure/azarm"
	"github.com/hexa-org/policy-orchestrator/pkg/azuretestsupport"
	"github.com/hexa-org/policy-orchestrator/pkg/azuretestsupport/apim_testsupport"
	"github.com/hexa-org/policy-orchestrator/pkg/azuretestsupport/armtestsupport"
	"github.com/hexa-org/policy-orchestrator/pkg/testsupport/policytestsupport"
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestGetPolicyInfo(t *testing.T) {
	apimApiSvc := apim_testsupport.NewMockArmApimSvc()
	azureClient := azuretestsupport.NewMockAzureClient()
	apimNamedValueSvc := apim_testsupport.NewMockApimNamedValueSvc()

	gatewayUrl := armtestsupport.ApimServiceGatewayUrl
	serviceInfo := apim_testsupport.ApimServiceInfo(gatewayUrl)
	apimApiSvc.ExpectGetApimServiceInfo(serviceInfo)
	apimApiSvc.ExpectGetApimServiceInfo(serviceInfo)
	existingActionRoles := map[string][]string{policytestsupport.ActionGetHrUs: {"some-role"}}
	apimNamedValueSvc.ExpectGetResourceRoles(serviceInfo, existingActionRoles)

	service := azarm.NewApimProviderService(apimApiSvc, azureClient, apimNamedValueSvc)
	appInfo := orchestrator.ApplicationInfo{Service: gatewayUrl}
	_, err := service.GetPolicyInfo(appInfo)
	assert.NoError(t, err)
}

func TestSetPolicyInfo_NoChange(t *testing.T) {
	apimApiSvc := apim_testsupport.NewMockArmApimSvc()
	azureClient := azuretestsupport.NewMockAzureClient()
	apimNamedValueSvc := apim_testsupport.NewMockApimNamedValueSvc()

	gatewayUrl := armtestsupport.ApimServiceGatewayUrl
	serviceInfo := apim_testsupport.ApimServiceInfo(gatewayUrl)
	apimApiSvc.ExpectGetApimServiceInfo(serviceInfo)
	existingActionRoles := map[string][]string{policytestsupport.ActionGetHrUs: {"some-role"}}
	apimNamedValueSvc.ExpectGetResourceRoles(serviceInfo, existingActionRoles)

	newActionRoles := map[string][]string{policytestsupport.ActionGetHrUs: {"some-role"}}
	policies := policytestsupport.MakeRoleSubjectTestPolicies(newActionRoles)
	service := azarm.NewApimProviderService(apimApiSvc, azureClient, apimNamedValueSvc)
	appInfo := orchestrator.ApplicationInfo{Service: gatewayUrl}
	_, err := service.SetPolicyInfo(appInfo, policies)

	assert.NoError(t, err)
}
