package azuretestsupport

import (
	"encoding/json"
	"github.com/Azure/azure-sdk-for-go/sdk/azcore"
	"github.com/hexa-org/policy-orchestrator/internal/orchestratorproviders/microsoftazure/azurecommon"
)

const AzureAppId = "anAppId"
const AzureAppName = "anAppName"
const AzureSubscription = "aSubscription"
const AzureTenantId = "aTenant"
const AzureSecret = "aSecret"

func AzureKey() azurecommon.AzureKey {
	return azurecommon.AzureKey{
		AppId:        AzureAppId,
		Secret:       AzureSecret,
		Tenant:       AzureTenantId,
		Subscription: AzureSubscription,
	}
}

func AzureTokenCredential(httpClient azurecommon.HTTPClient) azcore.TokenCredential {
	creds, _ := azurecommon.ClientSecretCredentials(AzureKey(), httpClient)
	return creds
}

func AzureKeyBytes() []byte {
	keyBytes, _ := json.Marshal(AzureKey())
	return keyBytes
}
