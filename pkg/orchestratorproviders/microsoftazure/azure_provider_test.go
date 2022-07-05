package microsoftazure_test

import (
	"github.com/hexa-org/policy-orchestrator/pkg/orchestrator"
	"github.com/hexa-org/policy-orchestrator/pkg/orchestratorproviders/microsoftazure"
	"github.com/hexa-org/policy-orchestrator/pkg/orchestratorproviders/microsoftazure/test"
	"github.com/hexa-org/policy-orchestrator/pkg/policysupport"
	"github.com/stretchr/testify/assert"
	"net/http"
	"testing"
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
      "description": "aDescription"
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
	assert.Equal(t, 1, len(applications))
	assert.Equal(t, "azure", p.Name())
}

func TestGetPolicy(t *testing.T) {
	m := new(microsoftazure_test.MockClient)
	m.Exchanges = []microsoftazure_test.MockExchange{
		{Path: "https://login.microsoftonline.com/aTenant/oauth2/v2.0/token", ResponseBody: []byte("{\"access_token\":\"aToken\"}")},
		{Path: "https://graph.microsoft.com/v1.0/servicePrincipals?$search=\"appId:aDescription\"", ResponseBody: []byte("{\"value\":[{\"id\":\"aToken\"}]}")},
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
	assert.Equal(t, "aPrincipalId:aPrincipalDisplayName", policies[0].Subject.Members[0])
	assert.Equal(t, "aResourceId:aResourceDisplayName", policies[0].Object.Resources[0])
}

func TestSetPolicy(t *testing.T) {
	m := new(microsoftazure_test.MockClient)
	m.Exchanges = []microsoftazure_test.MockExchange{
		{Path: "https://login.microsoftonline.com/aTenant/oauth2/v2.0/token", ResponseBody: []byte("{\"access_token\":\"aToken\"}")},
		{Path: "https://graph.microsoft.com/v1.0/servicePrincipals?$search=\"appId:aDescription\"", ResponseBody: []byte("{\"value\":[{\"id\":\"aToken\"}]}")},
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
	}

	azureProvider := microsoftazure.AzureProvider{HttpClientOverride: m}
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
	status, err := azureProvider.SetPolicyInfo(info, appInfo, []policysupport.PolicyInfo{{
		Actions: []policysupport.ActionInfo{{"azure:anAppRoleId"}},
		Subject: policysupport.SubjectInfo{Members: []string{"aPrincipalId:aPrincipalDisplayName", "yetAnotherPrincipalId:yetAnotherPrincipalDisplayName", "andAnotherPrincipalId:andAnotherPrincipalDisplayName"}},
		Object:  policysupport.ObjectInfo{Resources: []string{"aResourceId:aResourceDisplayName"}},
	}})
	assert.Equal(t, http.StatusCreated, status)
	assert.NoError(t, err)
}
