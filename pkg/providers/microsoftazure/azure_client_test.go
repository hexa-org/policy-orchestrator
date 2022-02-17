package microsoftazure_test

import (
	"errors"
	"github.com/hexa-org/policy-orchestrator/pkg/providers/microsoftazure"
	"github.com/hexa-org/policy-orchestrator/pkg/providers/microsoftazure/test"
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestAzureClient_GetWebApplications(t *testing.T) {
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
	client := microsoftazure.AzureClient{HttpClient: m}
	key := []byte(`{
	"appId":"anAppId",
	"secret":"aSecret",
	"tenant":"aTenant",
	"subscription":"aSubscription"
}`)
	applications, _ := client.GetWebApplications(key)
	assert.Equal(t, 1, len(applications))
	assert.Equal(t, "anAppName", applications[0].Name)
	assert.Equal(t, "app,linux,container", applications[0].Description)
}

func TestAzureClient_GetWebApplications_withBadKey(t *testing.T) {
	m := new(microsoftazure_test.MockClient)
	m.ResponseBody = []byte(`{"value": []}`)
	client := microsoftazure.AzureClient{HttpClient: m}
	_, err := client.GetWebApplications([]byte("aBadKey"))
	assert.Equal(t, "invalid character 'a' looking for beginning of value", err.Error())
}

func TestAzureClient_GetWebApplications_withRequestError(t *testing.T) {
	m := new(microsoftazure_test.MockClient)
	m.ResponseBody = []byte(`{"value": []}`)
	m.Err = errors.New("oops")

	client := microsoftazure.AzureClient{HttpClient: m}
	key := []byte(`{
	"appId":"anAppId",
	"secret":"aSecret",
	"tenant":"aTenant",
	"subscription":"aSubscription"
}`)
	_, err := client.GetWebApplications(key)
	assert.Equal(t, "oops", err.Error())
}

func TestAzureClient_GetWebApplications_withBadJson(t *testing.T) {
	m := new(microsoftazure_test.MockClient)
	m.ResponseBody = []byte(`_`)

	client := microsoftazure.AzureClient{HttpClient: m}
	key := []byte(`{
	"appId":"anAppId",
	"secret":"aSecret",
	"tenant":"aTenant",
	"subscription":"aSubscription"
}`)
	_, err := client.GetWebApplications(key)
	assert.Equal(t, "invalid character '_' looking for beginning of value", err.Error())
}
