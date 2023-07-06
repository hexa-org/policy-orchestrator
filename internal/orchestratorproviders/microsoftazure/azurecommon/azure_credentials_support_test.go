package azurecommon_test

import (
	"github.com/hexa-org/policy-orchestrator/internal/orchestratorproviders/microsoftazure/azurecommon"
	"github.com/hexa-org/policy-orchestrator/pkg/azuretestsupport"
	"github.com/stretchr/testify/assert"
	"net/http"
	"testing"
)

func TestClientSecretCredentials(t *testing.T) {
	tests := []struct {
		name   string
		client azurecommon.HTTPClient
	}{
		{name: "default client"},
		{name: "override client"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			azureKey := azuretestsupport.AzureKey()
			creds, err := azurecommon.ClientSecretCredentials(azureKey, &http.Client{})
			assert.NoError(t, err)
			assert.NotNil(t, creds)
		})
	}
}

func TestDecodeKey_Error(t *testing.T) {
	key, err := azurecommon.DecodeKey([]byte("x"))
	assert.ErrorContains(t, err, "'x'")
	assert.Empty(t, key)
}

func TestDecodeKey_Success(t *testing.T) {
	key, err := azurecommon.DecodeKey(azuretestsupport.AzureKeyBytes())
	assert.NoError(t, err)
	assert.Equal(t, azuretestsupport.AzureAppId, key.AppId)
	assert.Equal(t, azuretestsupport.AzureSecret, key.Secret)
	assert.Equal(t, azuretestsupport.AzureTenantId, key.Tenant)
	assert.Equal(t, azuretestsupport.AzureSubscription, key.Subscription)
}
