package microsoftazure_test

import (
	"errors"
	"github.com/hexa-org/policy-orchestrator/pkg/testsupport"
	"testing"

	"github.com/hexa-org/policy-orchestrator/pkg/orchestratorproviders/microsoftazure"
	"github.com/stretchr/testify/assert"
)

func TestAzureClient_GetWebApplications(t *testing.T) {
	m := testsupport.NewMockHTTPClient()
	m.ResponseBody["https://login.microsoftonline.com/aTenant/oauth2/v2.0/token"] = []byte("{\"access_token\":\"aToken\"}")
	m.ResponseBody["https://graph.microsoft.com/v1.0/applications"] = []byte(`
{
  "value": [
    {
      "id": "anObjectId",
      "appId": "anAppId",
      "displayName": "anAppName",
			"web": {
			  "homePageUrl": "https://anAppName.azurewebsites.net"
      }
    }
  ]
}
`)
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
	m := testsupport.NewMockHTTPClient()
	m.ResponseBody["https://login.microsoftonline.com/aTenant/oauth2/v2.0/token"] = []byte("{\"access_token\":\"aToken\"}")
	m.ResponseBody["https://graph.microsoft.com/v1.0/applications"] = []byte(`=P`)
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
	m := testsupport.NewMockHTTPClient()
	m.ResponseBody["https://login.microsoftonline.com/aTenant/oauth2/v2.0/token"] = []byte(`{"value": []}`)
	client := microsoftazure.AzureClient{HttpClient: m}

	_, err := client.GetWebApplications([]byte("aBadKey"))

	assert.Equal(t, "invalid character 'a' looking for beginning of value", err.Error())
}

func TestAzureClient_GetWebApplications_withRequestError(t *testing.T) {
	m := testsupport.NewMockHTTPClient()
	m.ResponseBody["https://login.microsoftonline.com/aTenant/oauth2/v2.0/token"] = []byte(`{"value": []}`)
	m.Err = errors.New("oops")
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
	m := testsupport.NewMockHTTPClient()
	m.ResponseBody["https://login.microsoftonline.com/aTenant/oauth2/v2.0/token"] = []byte("_")
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
	m := testsupport.NewMockHTTPClient()
	m.ResponseBody["https://login.microsoftonline.com/aTenant/oauth2/v2.0/token"] = []byte(`{"value": []}`)
	client := microsoftazure.AzureClient{HttpClient: m}

	_, err := client.GetServicePrincipals([]byte("aBadKey"), "")

	assert.Equal(t, "invalid character 'a' looking for beginning of value", err.Error())
}

func TestAzureClient_GetServicePrincipals_withBadPrincipalJson(t *testing.T) {
	m := testsupport.NewMockHTTPClient()
	m.ResponseBody["https://login.microsoftonline.com/aTenant/oauth2/v2.0/token"] = []byte("{\"access_token\":\"aToken\"}")
	m.ResponseBody["https://graph.microsoft.com/v1.0/servicePrincipals?$search=\"appId:anAppId\""] = []byte("~")

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

func TestAzureClient_GetUserInfoFromPrincipalId_withABadKey(t *testing.T) {
	m := testsupport.NewMockHTTPClient()
	m.ResponseBody["https://login.microsoftonline.com/aTenant/oauth2/v2.0/token"] = []byte(`{"value": []}`)
	client := microsoftazure.AzureClient{HttpClient: m}

	_, err := client.GetUserInfoFromPrincipalId([]byte("aBadKey"), "")

	assert.Equal(t, "invalid character 'a' looking for beginning of value", err.Error())
}

func TestAzureClient_GetUserInfoFromPrincipalId_withBadJson(t *testing.T) {
	m := testsupport.NewMockHTTPClient()
	m.ResponseBody["https://login.microsoftonline.com/aTenant/oauth2/v2.0/token"] = []byte("{\"access_token\":\"aToken\"}")
	m.ResponseBody["https://graph.microsoft.com/v1.0/users/aPrincipalId"] = []byte("~")

	client := microsoftazure.AzureClient{HttpClient: m}
	key := []byte(`
{
  "appId":"anAppId",
  "secret":"aSecret",
  "tenant":"aTenant",
  "subscription":"aSubscription"
}
`)

	_, err := client.GetUserInfoFromPrincipalId(key, "aPrincipalId")

	assert.Error(t, err)
}

func TestAzureClient_GetPrincipalIdFromEmail(t *testing.T) {
	m := testsupport.NewMockHTTPClient()
	m.ResponseBody["https://login.microsoftonline.com/aTenant/oauth2/v2.0/token"] = []byte("{\"access_token\":\"aToken\"}")
	m.ResponseBody["https://graph.microsoft.com/v1.0/users?$select=id,mail&$filter=mail%20eq%20%27anEmail%40example.com%27"] = []byte(
		"{\"value\":[{\"id\":\"anId\",\"mail\":\"anEmail@example.com\"}]}")

	client := microsoftazure.AzureClient{HttpClient: m}
	key := []byte(`
{
  "appId":"anAppId",
  "secret":"aSecret",
  "tenant":"aTenant",
  "subscription":"aSubscription"
}
`)

	response, _ := client.GetPrincipalIdFromEmail(key, "anEmail@example.com")

	assert.Equal(t, "anId", response)
}

func TestAzureClient_GetPrincipalIdFromEmail_withABadKey(t *testing.T) {
	m := testsupport.NewMockHTTPClient()
	m.ResponseBody["https://login.microsoftonline.com/aTenant/oauth2/v2.0/token"] = []byte(`{"value": []}`)
	client := microsoftazure.AzureClient{HttpClient: m}
	_, err := client.GetPrincipalIdFromEmail([]byte("aBadKey"), "")
	assert.Equal(t, "invalid character 'a' looking for beginning of value", err.Error())
}

func TestAzureClient_GetPrincipalIdFromEmail_withBadJson(t *testing.T) {
	m := testsupport.NewMockHTTPClient()
	m.ResponseBody["https://login.microsoftonline.com/aTenant/oauth2/v2.0/token"] = []byte("{\"access_token\":\"aToken\"}")
	m.ResponseBody["https://graph.microsoft.com/v1.0/users?$select=id,mail&$filter=mail%20eq%20%27anEmail%40example.com%27"] = []byte("~")

	client := microsoftazure.AzureClient{HttpClient: m}
	key := []byte(`
{
  "appId":"anAppId",
  "secret":"aSecret",
  "tenant":"aTenant",
  "subscription":"aSubscription"
}
`)

	_, err := client.GetPrincipalIdFromEmail(key, "anEmail@example.com")

	assert.Error(t, err)
}

func TestAzureClient_GetAppRoleAssignedTo_withABadKey(t *testing.T) {
	m := testsupport.NewMockHTTPClient()
	m.ResponseBody["https://login.microsoftonline.com/aTenant/oauth2/v2.0/token"] = []byte(`{"value": []}`)

	client := microsoftazure.AzureClient{HttpClient: m}
	_, err := client.GetAppRoleAssignedTo([]byte("aBadKey"), "")
	assert.Equal(t, "invalid character 'a' looking for beginning of value", err.Error())
}

func TestAzureClient_GetAppRoleAssignedTo_withBadJson(t *testing.T) {
	m := testsupport.NewMockHTTPClient()
	m.ResponseBody["https://login.microsoftonline.com/aTenant/oauth2/v2.0/token"] = []byte("{\"access_token\":\"aToken\"}")
	m.ResponseBody["https://graph.microsoft.com/v1.0/servicePrincipals/anAppId/appRoleAssignedTo"] = []byte("~")

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
	m := testsupport.NewMockHTTPClient()
	m.ResponseBody["https://login.microsoftonline.com/aTenant/oauth2/v2.0/token"] = []byte("{\"access_token\":\"aToken\"}")
	m.ResponseBody["https://graph.microsoft.com/v1.0/servicePrincipals?$search=\"appId:aDescription\""] = []byte("{\"value\":[{\"id\":\"aToken\"}]}")
	m.ResponseBody["https://graph.microsoft.com/v1.0/servicePrincipals/anAppId/appRoleAssignedTo"] = []byte(`
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
`)
	m.ResponseBody["https://graph.microsoft.com/v1.0/servicePrincipals/anAppId/appRoleAssignedTo/anId"] = []byte("")

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
		{ID: "anId"},
	})

	assert.NoError(t, err)
}

func TestAzureClient_SetAppRoleAssignedTo_withBadGet(t *testing.T) {
	m := testsupport.NewMockHTTPClient()
	m.ResponseBody["https://login.microsoftonline.com/aTenant/oauth2/v2.0/token"] = []byte("{\"access_token\":\"aToken\"}")
	m.ResponseBody["https://graph.microsoft.com/v1.0/servicePrincipals?$search=\"appId:aDescription\""] = []byte("{\"value\":[{\"id\":\"aToken\"}]}")
	m.ResponseBody["https://graph.microsoft.com/v1.0/servicePrincipals/anAppId/appRoleAssignedTo"] = []byte(`~`)

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
		{ID: "anId"},
	})
	assert.Error(t, err)
}

func TestAzureClient_SetAppRoleAssignedTo_withBadAdd(t *testing.T) {
	m := testsupport.NewMockHTTPClient()
	m.ResponseBody["https://login.microsoftonline.com/aTenant/oauth2/v2.0/token"] = []byte("{\"access_token\":\"aToken\"}")
	m.ResponseBody["https://graph.microsoft.com/v1.0/servicePrincipals?$search=\"appId:aDescription\""] = []byte("{\"value\":[{\"id\":\"aToken\"}]}")
	m.ResponseBody["https://graph.microsoft.com/v1.0/servicePrincipals/anAppId/appRoleAssignedTo"] = []byte(`
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
`)
	m.ResponseBody["https://graph.microsoft.com/v1.0/servicePrincipals/anAppId/appRoleAssignedTo/anId"] = []byte("{\"value\":[{\"id\":\"aToken\"}]}")
	m.Err = errors.New("oops")

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
		{ID: "anId"},
	})
	assert.EqualError(t, err, "oops")
}

func TestAzureClient_SetAppRoleAssignedTo_withBadDelete(t *testing.T) {
	//todo
}

func TestAzureClient_AddAppRolesAssignedTo(t *testing.T) {
	m := testsupport.NewMockHTTPClient()
	m.ResponseBody["https://login.microsoftonline.com/aTenant/oauth2/v2.0/token"] = []byte(`{"value": []}`)

	client := microsoftazure.AzureClient{HttpClient: m}
	err := client.AddAppRolesAssignedTo([]byte("aBadKey"), "", []microsoftazure.AzureAppRoleAssignment{
		{ID: "anId"},
	})
	assert.Equal(t, "invalid character 'a' looking for beginning of value", err.Error())
}

func TestAzureClient_DeleteAppRolesAssignedTo(t *testing.T) {
	m := testsupport.NewMockHTTPClient()
	m.ResponseBody["https://login.microsoftonline.com/aTenant/oauth2/v2.0/token"] = []byte(`{"value": []}`)
	client := microsoftazure.AzureClient{HttpClient: m}
	err := client.DeleteAppRolesAssignedTo([]byte("aBadKey"), "", []string{"anId"})
	assert.Equal(t, "invalid character 'a' looking for beginning of value", err.Error())
}

func TestAzureClient_ShouldAdd(t *testing.T) {
	m := testsupport.NewMockHTTPClient()
	m.ResponseBody["https://login.microsoftonline.com/aTenant/oauth2/v2.0/token"] = []byte(`{"value": []}`)
	client := microsoftazure.AzureClient{HttpClient: m}
	assignments := []microsoftazure.AzureAppRoleAssignment{{"anId", "anAppRoleId",
		"aPrincipalDisplayName", "aPrincipalId", "aPrincipalType",
		"aResourceDisplayName", "aResourceId"}}
	shouldAdd := client.ShouldAdd(assignments,
		microsoftazure.AzureAppRoleAssignments{List: []microsoftazure.AzureAppRoleAssignment{}})
	assert.Equal(t, 1, len(shouldAdd))
}

func TestAzureClient_ShouldRemove(t *testing.T) {
	m := testsupport.NewMockHTTPClient()
	m.ResponseBody["https://login.microsoftonline.com/aTenant/oauth2/v2.0/token"] = []byte(`{"value": []}`)
	client := microsoftazure.AzureClient{HttpClient: m}
	assignments := []microsoftazure.AzureAppRoleAssignment{{"anId", "anAppRoleId",
		"aPrincipalDisplayName", "aPrincipalId", "aPrincipalType",
		"aResourceDisplayName", "aResourceId"}}
	shouldAdd := client.ShouldRemove(
		microsoftazure.AzureAppRoleAssignments{List: []microsoftazure.AzureAppRoleAssignment{}},
		assignments)
	assert.Equal(t, 0, len(shouldAdd))
}
