package microsoftazure_test

import (
	"github.com/hexa-org/policy-orchestrator/pkg/testsupport"
	"net/http"
	"testing"

	"github.com/hexa-org/policy-orchestrator/pkg/orchestrator"
	"github.com/hexa-org/policy-orchestrator/pkg/orchestratorproviders/microsoftazure"
	"github.com/hexa-org/policy-orchestrator/pkg/policysupport"
	"github.com/stretchr/testify/assert"
)

func mockClientSetup() *testsupport.MockHTTPClient {
	m := testsupport.NewMockHTTPClient()
	m.ResponseBody["https://login.microsoftonline.com/aTenant/oauth2/v2.0/token"] = []byte("{\"access_token\":\"aToken\"}")
	m.ResponseBody["https://graph.microsoft.com/v1.0/servicePrincipals?$search=\"appId:aDescription\""] = []byte("{\"value\":[{\"id\":\"aToken\"}]}")
	m.ResponseBody["https://graph.microsoft.com/v1.0/users/aPrincipalId"] = []byte("{\"mail\":\"anEmail@example.com\"}")
	m.ResponseBody["https://graph.microsoft.com/v1.0/servicePrincipals/aToken/appRoleAssignedTo"] = []byte(`
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
`)
	return m
}

func TestDiscoverApplications(t *testing.T) {
	m := testsupport.NewMockHTTPClient()
	m.ResponseBody["https://login.microsoftonline.com/aTenant/oauth2/v2.0/token"] = []byte("{\"access_token\":\"aToken\"}")
	m.ResponseBody["https://graph.microsoft.com/v1.0/applications"] = []byte(`
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
}`)

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
	m := mockClientSetup()
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

func TestGetPolicy_withOutAUserEmail(t *testing.T) {
	m := mockClientSetup()
	m.ResponseBody["https://graph.microsoft.com/v1.0/users/aPrincipalId"] = []byte("{\"mail\":\"\"")

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
	assert.Empty(t, policies[0].Subject.Members, "empty")
	assert.Equal(t, "anObjectId", policies[0].Object.ResourceID)
}

func TestSetPolicy(t *testing.T) {
	m := mockClientSetup()
	m.ResponseBody["https://graph.microsoft.com/v1.0/servicePrincipals/aToken/appRoleAssignedTo/anId"] = []byte("")
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
	m := testsupport.NewMockHTTPClient()
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
		orchestrator.ApplicationInfo{Name: "anAppName", Description: "aDescription"},
		[]policysupport.PolicyInfo{{
			Meta:    policysupport.MetaInfo{Version: "0"},
			Actions: []policysupport.ActionInfo{{"azure:anAppRoleId"}},
			Subject: policysupport.SubjectInfo{Members: []string{"aPrincipalId:aPrincipalDisplayName", "yetAnotherPrincipalId:yetAnotherPrincipalDisplayName", "andAnotherPrincipalId:andAnotherPrincipalDisplayName"}},
			Object: policysupport.ObjectInfo{
				ResourceID: "anObjectId",
			},
		}})

	assert.Equal(t, http.StatusInternalServerError, status)
	assert.EqualError(t, err, "Key: 'ApplicationInfo.ObjectID' Error:Field validation for 'ObjectID' failed on the 'required' tag")

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
	assert.EqualError(t, err, "Key: '[0].Object.ResourceID' Error:Field validation for 'ResourceID' failed on the 'required' tag")
}
