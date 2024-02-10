package azurecommon_test

import (
	"encoding/json"
	"github.com/Azure/azure-sdk-for-go/sdk/azcore"
	"github.com/hexa-org/policy-orchestrator/sdk/providerazure/azurecommon"
	"github.com/stretchr/testify/assert"
	"net/http"
	"testing"
)

const AzureAppId = "anAppId"
const AzureAppName = "anAppName"
const AzureSubscription = "aSubscription"
const AzureTenantId = "aTenant"
const AzureSecret = "aSecret"

func TestClientSecretCredentials(t *testing.T) {
	tests := []struct {
		name   string
		client azurecommon.HTTPClient
		badKey bool
	}{
		{name: "default client"},
		{name: "override client", client: &http.Client{}},
		{name: "bad key", badKey: true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			key := AzureKeyBytes()
			if tt.badKey {
				key = []byte("~")
			}
			creds, err := azurecommon.ClientSecretCredentials(key, tt.client)
			assert.Equal(t, tt.badKey, err != nil)
			assert.Equal(t, tt.badKey, creds == nil)
		})
	}
}

func TestDecodeKey_Error(t *testing.T) {
	key, err := azurecommon.DecodeKey([]byte("x"))
	assert.ErrorContains(t, err, "'x'")
	assert.Empty(t, key)
}

func TestDecodeKey_Success(t *testing.T) {
	key, err := azurecommon.DecodeKey(AzureKeyBytes())
	assert.NoError(t, err)
	assert.Equal(t, AzureAppId, key.AppId)
	assert.Equal(t, AzureSecret, key.Secret)
	assert.Equal(t, AzureTenantId, key.Tenant)
	assert.Equal(t, AzureSubscription, key.Subscription)
}

func AzureKey() azurecommon.AzureKey {
	return azurecommon.AzureKey{
		AppId:        AzureAppId,
		Secret:       AzureSecret,
		Tenant:       AzureTenantId,
		Subscription: AzureSubscription,
	}
}

func AzureTokenCredential(httpClient azurecommon.HTTPClient) azcore.TokenCredential {
	creds, _ := azurecommon.ClientSecretCredentials(AzureKeyBytes(), httpClient)
	return creds
}

func AzureKeyBytes() []byte {
	keyBytes, _ := json.Marshal(AzureKey())
	return keyBytes
}
