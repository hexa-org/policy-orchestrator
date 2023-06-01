package azuretestsupport

import (
	"encoding/json"
	"fmt"
	"github.com/hexa-org/policy-orchestrator/internal/orchestratorproviders/microsoftazure"
	"github.com/hexa-org/policy-orchestrator/pkg/testsupport/policytestsupport"
)

const ServicePrincipalId = "some-service-principal-id"
const AzureAppId = "anAppId"
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

func AzureClientKey() []byte {
	key := microsoftazure.AzureKey{
		AppId:        AzureAppId,
		Secret:       "aSecret",
		Tenant:       AzureTenantId,
		Subscription: "aSubscription",
	}

	keyBytes, _ := json.Marshal(key)
	return keyBytes
}

func AzureServicePrincipals() microsoftazure.AzureServicePrincipals {
	var sps microsoftazure.AzureServicePrincipals
	_ = json.Unmarshal([]byte(ServicePrincipalsRespJson), &sps)
	return sps
}
