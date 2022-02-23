package microsoftazure_test

import (
	"errors"
	"github.com/hexa-org/policy-orchestrator/pkg/orchestrator/provider"
	"github.com/hexa-org/policy-orchestrator/pkg/providers/microsoftazure"
	"github.com/hexa-org/policy-orchestrator/pkg/providers/microsoftazure/test"
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestAzureClient_GetWebApplications(t *testing.T) {
	m := new(microsoftazure_test.MockClient)
	m.Exchanges = []microsoftazure_test.MockExchange{
		{Path: "https://login.microsoftonline.com/aTenant/oauth2/v2.0/token", ResponseBody: []byte("{\"access_token\":\"aToken\"}")},
		{Path: "https://graph.microsoft.com/v1.0/applications", ResponseBody: []byte(`
{
  "value": [
    {
      "id": "anObjectId",
      "appId": "anAppId",
      "displayName": "anAppName"
    }
  ]
}
`),
		},
	}
	client := microsoftazure.AzureClient{HttpClient: m}
	key := []byte(`
{
  "appId":"anAppId",
  "secret":"aSecret",
  "tenant":"aTenant",
  "subscription":"aSubscription"
}
`)
	applications, _ := client.GetWebApplications(key)
	assert.Equal(t, 1, len(applications))
	assert.Equal(t, "anAppName", applications[0].Name)
	assert.Equal(t, "anAppId", applications[0].Description)
}

func TestAzureClient_GetWebApplications_withBadKey(t *testing.T) {
	client := clientForTesting()
	_, err := client.GetWebApplications([]byte("aBadKey"))
	assert.Equal(t, "invalid character 'a' looking for beginning of value", err.Error())
}

func TestAzureClient_GetWebApplications_withRequestError(t *testing.T) {
	m := new(microsoftazure_test.MockClient)
	m.Exchanges = []microsoftazure_test.MockExchange{
		{Path: "https://login.microsoftonline.com/aTenant/oauth2/v2.0/token", ResponseBody: []byte(`{"value": []}`), Err: errors.New("oops")},
	}
	client := microsoftazure.AzureClient{HttpClient: m}
	key := []byte(`
{
  "appId":"anAppId",
  "secret":"aSecret",
  "tenant":"aTenant",
  "subscription":"aSubscription"
}
`)
	_, err := client.GetWebApplications(key)
	assert.Equal(t, "oops", err.Error())
}

func TestAzureClient_GetWebApplications_withBadJson(t *testing.T) {
	m := new(microsoftazure_test.MockClient)
	m.Exchanges = []microsoftazure_test.MockExchange{
		{Path: "https://login.microsoftonline.com/aTenant/oauth2/v2.0/token", ResponseBody: []byte("_")},
	}

	client := microsoftazure.AzureClient{HttpClient: m}
	key := []byte(`
{
  "appId":"anAppId",
  "secret":"aSecret",
  "tenant":"aTenant",
  "subscription":"aSubscription"
}
`)
	_, err := client.GetWebApplications(key)
	assert.Equal(t, "invalid character '_' looking for beginning of value", err.Error())
}

func TestAzureClient_GetServicePrincipals_withABadKey(t *testing.T) {
	client := clientForTesting()
	_, err := client.GetServicePrincipals([]byte("aBadKey"), "")
	assert.Equal(t, "invalid character 'a' looking for beginning of value", err.Error())
}

func TestAzureClient_GetAppRoleAssignedTo_withABadKey(t *testing.T) {
	client := clientForTesting()
	_, err := client.GetAppRoleAssignedTo([]byte("aBadKey"), "")
	assert.Equal(t, "invalid character 'a' looking for beginning of value", err.Error())
}

func TestAzureClient_SetPolicy(t *testing.T) {
	m := new(microsoftazure_test.MockClient)

	client := microsoftazure.AzureClient{HttpClient: m}
	key := []byte(`
{
  "appId":"anAppId",
  "secret":"aSecret",
  "tenant":"aTenant",
  "subscription":"aSubscription"
}
`)
	_ = client.SetPolicy(key, provider.PolicyInfo{})
}

func clientForTesting() microsoftazure.AzureClient {
	m := new(microsoftazure_test.MockClient)
	m.Exchanges = []microsoftazure_test.MockExchange{
		{Path: "https://login.microsoftonline.com/aTenant/oauth2/v2.0/token", ResponseBody: []byte(`{"value": []}`)},
	}
	return microsoftazure.AzureClient{HttpClient: m}
}
