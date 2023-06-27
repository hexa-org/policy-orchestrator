package azuretestsupport

import (
	"encoding/json"
	"fmt"
	"github.com/hexa-org/policy-orchestrator/internal/orchestratorproviders/microsoftazure/azad"
	"github.com/hexa-org/policy-orchestrator/internal/orchestratorproviders/microsoftazure/azurecommon"
	"github.com/hexa-org/policy-orchestrator/pkg/testsupport/policytestsupport"
)

const AzureSubscription = "aSubscription"
const ServicePrincipalId = "some-service-principal-id"
const AzureAppId = "anAppId"
const AzureAppName = "anAppName"
const AzureTenantId = "aTenant"

type AppRoleId string

const AppRoleIdGetHrUs AppRoleId = "app-role-get-hr-us"
const AppRoleIdGetProfile AppRoleId = "app-role-get-profile"

var ServicePrincipalsRespJson = fmt.Sprintf(`{"value": [
{
	"id": "%s",
	"displayName": "%s",
	"appRoles": [
		{
			"allowedMemberTypes": [
				"User"
			],
			"description": "Allows GET to the /humanresources/us",
			"displayName": "GetHR-US",
			"id": "%s",
			"isEnabled": true,
			"origin": "Application",
			"value": "%s"
		},
		{
			"allowedMemberTypes": [
				"User"
			],
			"description": "Allows GET to the /profile",
			"displayName": "AppRoleIdGetProfile",
			"id": "%s",
			"isEnabled": true,
			"origin": "Application",
			"value": "%s"
		}
	] 
}]}`, ServicePrincipalId, policytestsupport.PolicyObjectResourceId, AppRoleIdGetHrUs, policytestsupport.ActionGetHrUs, AppRoleIdGetProfile, policytestsupport.ActionGetProfile)

func AzureKey() azurecommon.AzureKey {
	return azurecommon.AzureKey{
		AppId:        AzureAppId,
		Secret:       "aSecret",
		Tenant:       AzureTenantId,
		Subscription: AzureSubscription,
	}
}

func AzureClientKey() []byte {
	keyBytes, _ := json.Marshal(AzureKey())
	return keyBytes
}

func AzureServicePrincipals() azad.AzureServicePrincipals {
	var sps azad.AzureServicePrincipals
	_ = json.Unmarshal([]byte(ServicePrincipalsRespJson), &sps)
	return sps
}
