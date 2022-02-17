package microsoftazure_test

import (
	"github.com/hexa-org/policy-orchestrator/pkg/orchestrator/provider"
	"github.com/hexa-org/policy-orchestrator/pkg/providers/microsoftazure"
	"github.com/hexa-org/policy-orchestrator/pkg/providers/microsoftazure/test"
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestDiscoverApplications(t *testing.T) {
	m := new(microsoftazure_test.MockClient)
	m.ResponseBody = []byte(`{
  "value": [
    {
      "id": "/subscriptions/aSubscription/resourceGroups/aResourceGroup/providers/Microsoft.Web/sites/azurepullrequest",
      "name": "anAppName",
      "type": "Microsoft.Web/serverfarms/sites",
      "kind": "app,linux,container"
	}]
}`)
	providers := []provider.Provider{microsoftazure.AzureProvider{Http: m}}
	key := []byte(`{
	"appId":"anAppId",
	"secret":"aSecret",
	"tenant":"aTenant",
	"subscription":"aSubscription"
}`)
	for _, p := range providers {
		info := provider.IntegrationInfo{Name: "azure", Key: []byte(key)}
		applications, _ := p.DiscoverApplications(info)
		assert.Equal(t, 1, len(applications))
		assert.Equal(t, "azure", p.Name())
	}
}

func TestGetPolicy(t *testing.T) {
	m := new(microsoftazure_test.MockClient)
	p := microsoftazure.AzureProvider{Http: m}
	_, err := p.GetPolicyInfo(provider.IntegrationInfo{}, provider.ApplicationInfo{})
	assert.NoError(t, err)
}

func TestSetPolicy(t *testing.T) {
	m := new(microsoftazure_test.MockClient)
	p := microsoftazure.AzureProvider{Http: m}
	err := p.SetPolicyInfo(provider.IntegrationInfo{}, provider.ApplicationInfo{}, provider.PolicyInfo{})
	assert.NoError(t, err)
}
