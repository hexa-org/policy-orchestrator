package microsoftazure_test

import (
	"net/http"
	"testing"

	"github.com/hexa-org/policy-orchestrator/pkg/orchestrator"
	"github.com/hexa-org/policy-orchestrator/pkg/orchestratorproviders/microsoftazure"
	"github.com/hexa-org/policy-orchestrator/pkg/orchestratorproviders/microsoftazure/test"
	"github.com/hexa-org/policy-orchestrator/pkg/policysupport"
	"github.com/stretchr/testify/assert"
)

func TestDiscoverApplications(t *testing.T) {
	m := new(microsoftazure_test.MockClient)
	m.Exchanges = []microsoftazure_test.MockExchange{
		{Path: "https://login.microsoftonline.com/aTenant/oauth2/v2.0/token", ResponseBody: []byte("{\"access_token\":\"aToken\"}")},
		{Path: "https://graph.microsoft.com/v1.0/applications", ResponseBody: []byte(`
{
  "value": [
    {
      "id": "anId",
      "name": "anAppName",
      "description": "aDescription",
      "web": {
		"homePageUrl": "https://anAppName.azurewebsites.net"
      }
    }
  ]
}
`)}}

	p := &microsoftazure.AzureProvider{HttpClientOverride: m}
	key := []byte(`
{
  "appId":"anAppId",
  "secret":"aSecret",
  "tenant":"aTenant",
  "subscription":"aSubscription"
}
`)
	info := orchestrator.IntegrationInfo{Name: "azure", Key: key}
	applications, _ := p.DiscoverApplications(info)
	assert.Len(t, applications, 1)
	assert.Equal(t, "azure", p.Name())
	assert.Equal(t, "App Service", applications[0].Service)
}

func TestGetPolicy(t *testing.T) {
	m := new(microsoftazure_test.MockClient)
	m.Exchanges = []microsoftazure_test.MockExchange{
		{Path: "https://login.microsoftonline.com/aTenant/oauth2/v2.0/token", ResponseBody: []byte("{\"access_token\":\"aToken\"}")},
		{Path: "https://graph.microsoft.com/v1.0/servicePrincipals?$search=\"appId:aDescription\"", ResponseBody: []byte("{\"value\":[{\"id\":\"aToken\"}]}")},
		{Path: "https://graph.microsoft.com/v1.0/users/aPrincipalId", ResponseBody: []byte("{\"mail\":\"anEmail@example.com\"}")},
		{Path: "https://graph.microsoft.com/v1.0/servicePrincipals/aToken/appRoleAssignedTo", ResponseBody: []byte(`
{
  "value": [
    {
      "id": "anId",
      "appRoleId": "anAppRoleId",
      "principalDisplayName": "aPrincipalDisplayName",
      "principalId": "aPrincipalId",
      "principalType": "aPrincipalType",
      "resourceDisplayName": "aResourceDisplayName",
      "resourceId": "aResourceId"
    }
  ]
}
`)}}

	p := &microsoftazure.AzureProvider{HttpClientOverride: m}
	key := []byte(`
{
  "appId":"anAppId",
  "secret":"aSecret",
  "tenant":"aTenant",
  "subscription":"aSubscription"
}
`)
	info := orchestrator.IntegrationInfo{Name: "azure", Key: key}
	appInfo := orchestrator.ApplicationInfo{ObjectID: "anObjectId", Name: "anAppName", Description: "aDescription"}
	policies, _ := p.GetPolicyInfo(info, appInfo)
	assert.Equal(t, 1, len(policies))
	assert.Equal(t, "azure:anAppRoleId", policies[0].Actions[0].ActionUri)
	assert.Equal(t, "user:anEmail@example.com", policies[0].Subject.Members[0])
	assert.Equal(t, "anObjectId", policies[0].Object.ResourceID)
}

func TestSetPolicy(t *testing.T) {
	m := new(microsoftazure_test.MockClient)
	mockExchanges(m)
	azureProvider := microsoftazure.AzureProvider{HttpClientOverride: m}
	key := []byte(`
{
  "appId":"anAppId",
  "secret":"aSecret",
  "tenant":"aTenant",
  "subscription":"aSubscription"
}
`)
	status, err := azureProvider.SetPolicyInfo(
		orchestrator.IntegrationInfo{Name: "azure", Key: key},
		orchestrator.ApplicationInfo{ObjectID: "anObjectId", Name: "anAppName", Description: "aDescription"},
		[]policysupport.PolicyInfo{{
			Meta:    policysupport.MetaInfo{Version: "0"},
			Actions: []policysupport.ActionInfo{{"azure:anAppRoleId"}},
			Subject: policysupport.SubjectInfo{Members: []string{"user:anEmail@example.com"}},
			Object: policysupport.ObjectInfo{
				ResourceID: "anObjectId",
			},
		}})

	assert.Equal(t, http.StatusCreated, status)
	assert.NoError(t, err)
}

func TestSetPolicy_withInvalidArguments(t *testing.T) {
	m := new(microsoftazure_test.MockClient)
	mockExchanges(m)
	azureProvider := microsoftazure.AzureProvider{HttpClientOverride: m}
	key := []byte(`
{
  "appId":"anAppId",
  "secret":"aSecret",
  "tenant":"aTenant",
  "subscription":"aSubscription"
}
`)
	status, err := azureProvider.SetPolicyInfo(
		orchestrator.IntegrationInfo{Name: "azure", Key: key},
		orchestrator.ApplicationInfo{Name: "anAppName", Description: "aDescription"}, // missing objectId
		[]policysupport.PolicyInfo{{
			Meta:    policysupport.MetaInfo{Version: "0"},
			Actions: []policysupport.ActionInfo{{"azure:anAppRoleId"}},
			Subject: policysupport.SubjectInfo{Members: []string{"aPrincipalId:aPrincipalDisplayName", "yetAnotherPrincipalId:yetAnotherPrincipalDisplayName", "andAnotherPrincipalId:andAnotherPrincipalDisplayName"}},
			Object: policysupport.ObjectInfo{
				ResourceID: "anObjectId",
			},
		}})
	assert.Equal(t, http.StatusInternalServerError, status)
	assert.Error(t, err)

	status, err = azureProvider.SetPolicyInfo(
		orchestrator.IntegrationInfo{Name: "azure", Key: key},
		orchestrator.ApplicationInfo{ObjectID: "anObjectId", Name: "anAppName", Description: "aDescription"},
		[]policysupport.PolicyInfo{{
			Meta:    policysupport.MetaInfo{Version: "0"},
			Actions: []policysupport.ActionInfo{{"azure:anAppRoleId"}},
			Subject: policysupport.SubjectInfo{Members: []string{"aPrincipalId:aPrincipalDisplayName", "yetAnotherPrincipalId:yetAnotherPrincipalDisplayName", "andAnotherPrincipalId:andAnotherPrincipalDisplayName"}},
			Object:  policysupport.ObjectInfo{},
		}})
	assert.Equal(t, http.StatusInternalServerError, status)
	assert.Error(t, err)
}

func mockExchanges(m *microsoftazure_test.MockClient) {
	m.Exchanges = []microsoftazure_test.MockExchange{
		{Path: "https://login.microsoftonline.com/aTenant/oauth2/v2.0/token", ResponseBody: []byte("{\"access_token\":\"aToken\"}")},
		{Path: "https://graph.microsoft.com/v1.0/servicePrincipals?$search=\"appId:aDescription\"", ResponseBody: []byte("{\"value\":[{\"id\":\"aToken\"}]}")},
		{Path: "https://graph.microsoft.com/v1.0/users?$select=id,mail&$filter=mail%20eq%20%27anEmail@example.com%27",
			ResponseBody: []byte("{\"value\":[{\"id\":\"anId\",\"mail\":\"anEmail@example.com\"}]}")},
		{Path: "https://graph.microsoft.com/v1.0/servicePrincipals/aToken/appRoleAssignedTo", ResponseBody: []byte(`
{
  "value": [
    {
      "id": "anId",
      "appRoleId": "anAppRoleId", 
      "principalId": "aPrincipalId",
      "principalDisplayName": "aPrincipalDisplayName",
      "resourceId": "aResourceId",
      "resourceDisplayName": "aResourceDisplayName"
    },{
      "id": "anotherId",
      "appRoleId": "anotherAppRoleId", 
      "principalId": "anotherPrincipalId",
      "principalDisplayName": "anotherPrincipalDisplayName",
      "resourceId": "anotherResourceId",
      "resourceDisplayName": "anotherResourceDisplayName"
    },{
      "id": "andAnotherId",
      "appRoleId": "andAnotherAppRoleId", 
      "principalId": "andAnotherPrincipalId",
      "principalDisplayName": "andAnotherPrincipalDisplayName",
      "resourceId": "andAnotherResourceId",
      "resourceDisplayName": "andAnotherResourceDisplayName"
    }
  ]
}
`)},
		{Path: "https://graph.microsoft.com/v1.0/servicePrincipals/aToken/appRoleAssignedTo/anotherId"},
		{Path: "https://graph.microsoft.com/v1.0/servicePrincipals/aToken/appRoleAssignedTo/anId"},
	}
}
