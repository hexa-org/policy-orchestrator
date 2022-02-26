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

func TestAzureClient_GetWebApplications_withBadAppJson(t *testing.T) {
	m := new(microsoftazure_test.MockClient)
	m.Exchanges = []microsoftazure_test.MockExchange{
		{Path: "https://login.microsoftonline.com/aTenant/oauth2/v2.0/token", ResponseBody: []byte("{\"access_token\":\"aToken\"}")},
		{Path: "https://graph.microsoft.com/v1.0/applications", ResponseBody: []byte(`_`)},
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
	assert.Error(t, err)
}

func TestAzureClient_GetWebApplications_withABadKey(t *testing.T) {
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

func TestAzureClient_GetWebApplications_withBadJsonToken(t *testing.T) {
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

func TestAzureClient_GetServicePrincipals_withBadPrincipalJson(t *testing.T) {
	m := new(microsoftazure_test.MockClient)
	m.Exchanges = []microsoftazure_test.MockExchange{
		{Path: "https://login.microsoftonline.com/aTenant/oauth2/v2.0/token", ResponseBody: []byte("{\"access_token\":\"aToken\"}")},
		{Path: "https://graph.microsoft.com/v1.0/servicePrincipals?$search=\"appId:anAppId\"", ResponseBody: []byte("~")},
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
	_, err := client.GetServicePrincipals(key, "anAppId")
	assert.Error(t, err)
}

func TestAzureClient_GetAppRoleAssignedTo_withABadKey(t *testing.T) {
	client := clientForTesting()
	_, err := client.GetAppRoleAssignedTo([]byte("aBadKey"), "")
	assert.Equal(t, "invalid character 'a' looking for beginning of value", err.Error())
}

func TestAzureClient_GetAppRoleAssignedTo_withBadJson(t *testing.T) {
	m := new(microsoftazure_test.MockClient)
	m.Exchanges = []microsoftazure_test.MockExchange{
		{Path: "https://login.microsoftonline.com/aTenant/oauth2/v2.0/token", ResponseBody: []byte("{\"access_token\":\"aToken\"}")},
		{Path: "https://graph.microsoft.com/v1.0/servicePrincipals/anAppId/appRoleAssignedTo", ResponseBody: []byte("~")},
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
	_, err := client.GetAppRoleAssignedTo(key, "anAppId")
	assert.Error(t, err)
}

func TestAzureClient_SetAppRoleAssignedTo(t *testing.T) {
	m := new(microsoftazure_test.MockClient)
	m.Exchanges = []microsoftazure_test.MockExchange{
		{Path: "https://login.microsoftonline.com/aTenant/oauth2/v2.0/token", ResponseBody: []byte("{\"access_token\":\"aToken\"}")},
		{Path: "https://graph.microsoft.com/v1.0/servicePrincipals?$search=\"appId:aDescription\"", ResponseBody: []byte("{\"value\":[{\"id\":\"aToken\"}]}")},
		{Path: "https://graph.microsoft.com/v1.0/servicePrincipals/anAppId/appRoleAssignedTo", ResponseBody: []byte(`
{
  "value": [
    {
      "id": "anId",
      "appRoleId": "anAppRoleId", 
      "principalId": "aPrincipalId",
      "principalDisplayName": "aPrincipalDisplayName",
      "resourceId": "aResourceId",
      "resourceDisplayName": "aResourceDisplayName"
    }
  ]
}
`)},
		{Path: "https://graph.microsoft.com/v1.0/servicePrincipals/anAppId/appRoleAssignedTo/anId"},
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
	err := client.SetAppRoleAssignedTo(key, "anAppId", []microsoftazure.AzureAppRoleAssignment{
		{ID:"anId"},
	})
	assert.NoError(t, err)
}

func TestAzureClient_SetAppRoleAssignedTo_withBadGet(t *testing.T) {
	m := new(microsoftazure_test.MockClient)
	m.Exchanges = []microsoftazure_test.MockExchange{
		{Path: "https://login.microsoftonline.com/aTenant/oauth2/v2.0/token", ResponseBody: []byte("{\"access_token\":\"aToken\"}")},
		{Path: "https://graph.microsoft.com/v1.0/servicePrincipals?$search=\"appId:aDescription\"", ResponseBody: []byte("{\"value\":[{\"id\":\"aToken\"}]}")},
		{Path: "https://graph.microsoft.com/v1.0/servicePrincipals/anAppId/appRoleAssignedTo", ResponseBody: []byte(`~`)},
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
	err := client.SetAppRoleAssignedTo(key, "anAppId", []microsoftazure.AzureAppRoleAssignment{
		{ID:"anId"},
	})
	assert.Error(t, err)
}

func TestAzureClient_SetAppRoleAssignedTo_withBadAdd(t *testing.T) {
	m := new(microsoftazure_test.MockClient)
	m.Exchanges = []microsoftazure_test.MockExchange{
		{Path: "https://login.microsoftonline.com/aTenant/oauth2/v2.0/token", ResponseBody: []byte("{\"access_token\":\"aToken\"}")},
		{Path: "https://graph.microsoft.com/v1.0/servicePrincipals?$search=\"appId:aDescription\"", ResponseBody: []byte("{\"value\":[{\"id\":\"aToken\"}]}")},
		{Path: "https://graph.microsoft.com/v1.0/servicePrincipals/anAppId/appRoleAssignedTo", ResponseBody: []byte(`
{
  "value": [
    {
      "id": "anId",
      "appRoleId": "anAppRoleId", 
      "principalId": "aPrincipalId",
      "principalDisplayName": "aPrincipalDisplayName",
      "resourceId": "aResourceId",
      "resourceDisplayName": "aResourceDisplayName"
    }
  ]
}
`)},
		{Path: "https://graph.microsoft.com/v1.0/servicePrincipals/anAppId/appRoleAssignedTo/anId", Err: errors.New("oops")},
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
	err := client.SetAppRoleAssignedTo(key, "anAppId", []microsoftazure.AzureAppRoleAssignment{
		{ID:"anId"},
	})
	assert.Error(t, err)
}

func TestAzureClient_SetAppRoleAssignedTo_withBadDelete(t *testing.T) {

}

func TestAzureClient_AddAppRolesAssignedTo(t *testing.T) {
	client := clientForTesting()
	err := client.AddAppRolesAssignedTo([]byte("aBadKey"), "", []microsoftazure.AzureAppRoleAssignment{
		{ID:"anId"},
	})
	assert.Equal(t, "invalid character 'a' looking for beginning of value", err.Error())
}

func TestAzureClient_DeleteAppRolesAssignedTo(t *testing.T) {
	client := clientForTesting()
	err := client.DeleteAppRolesAssignedTo([]byte("aBadKey"), "", []string{"anId"})
	assert.Equal(t, "invalid character 'a' looking for beginning of value", err.Error())
}

func TestAzureClient_ShouldAdd(t *testing.T) {
	client := clientForTesting()
	assignments := []microsoftazure.AzureAppRoleAssignment{{"anId", "anAppRoleId",
		"aPrincipalDisplayName", "aPrincipalId", "aPrincipalType",
		"aResourceDisplayName", "aResourceId"}}
	shouldAdd := client.ShouldAdd(assignments,
		microsoftazure.AzureAppRoleAssignments{List: []microsoftazure.AzureAppRoleAssignment{}})
	assert.Equal(t, 1, len(shouldAdd))
}

func TestAzureClient_ShouldRemove(t *testing.T) {
	client := clientForTesting()
	assignments := []microsoftazure.AzureAppRoleAssignment{{"anId", "anAppRoleId",
		"aPrincipalDisplayName", "aPrincipalId", "aPrincipalType",
		"aResourceDisplayName", "aResourceId"}}
	shouldAdd := client.ShouldRemove(
		microsoftazure.AzureAppRoleAssignments{List: []microsoftazure.AzureAppRoleAssignment{}},
		assignments)
	assert.Equal(t, 0, len(shouldAdd))
}

func clientForTesting() microsoftazure.AzureClient {
	m := new(microsoftazure_test.MockClient)
	m.Exchanges = []microsoftazure_test.MockExchange{
		{Path: "https://login.microsoftonline.com/aTenant/oauth2/v2.0/token", ResponseBody: []byte(`{"value": []}`)},
	}
	return microsoftazure.AzureClient{HttpClient: m}
}
