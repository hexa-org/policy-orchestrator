package microsoftazure_test

import (
	"github.com/hexa-org/policy-orchestrator/internal/orchestratorproviders/microsoftazure"
	"github.com/hexa-org/policy-orchestrator/pkg/testsupport/azuretestsupport"
	"github.com/hexa-org/policy-orchestrator/pkg/testsupport/policytestsupport"
	"net/http"
	"testing"

	"github.com/stretchr/testify/assert"
)

var ExpAppsAzureResp = `
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
`

func TestAzureClient_GetWebApplications_Errors(t *testing.T) {
	m := azuretestsupport.NewAzureHttpClient()
	tests := []struct {
		name        string
		badToken    bool
		badKey      bool
		badRequest  bool
		badResponse bool
		errSubstr   string
	}{
		{
			name:      "Bad Key",
			badKey:    true,
			errSubstr: "invalid character 'k'",
		},
		{
			name:      "Bad Access Token",
			badToken:  true,
			errSubstr: "invalid character 'a'",
		},
		{
			name:       "Unexpected Status",
			badRequest: true,
			errSubstr:  "Unexpected status",
		},
		{
			name:        "Bad Response",
			badResponse: true,
			errSubstr:   "invalid character '~'",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			accessToken := "accessToken"
			key := azuretestsupport.AzureClientKey()

			if tt.badKey {
				key = []byte("keyInvalid")
			}

			if tt.badToken {
				accessToken = `"accessToken"`
			}

			m.TokenRequest(accessToken)
			if tt.badRequest {
				m.ErrorRequest(http.MethodGet, azuretestsupport.GraphApiBaseUrl+"/applications", http.StatusBadRequest, nil)
			}

			if tt.badResponse {
				m.GetWebApplicationsRequest("~")
			}

			client := m.AzureClient()
			apps, err := client.GetWebApplications(key)
			assert.Error(t, err)
			assert.ErrorContains(t, err, tt.errSubstr)
			assert.NotNil(t, apps)
			assert.Empty(t, apps)
		})
	}
}

func TestAzureClient_GetWebApplications(t *testing.T) {
	m := azuretestsupport.NewAzureHttpClient()

	m.TokenRequest("accessToken")
	m.GetWebApplicationsRequest(ExpAppsAzureResp)

	client := m.AzureClient()
	applications, err := client.GetWebApplications(azuretestsupport.AzureClientKey())
	assert.NoError(t, err)
	assert.NotNil(t, applications)
	assert.Equal(t, 1, len(applications))
	assert.Equal(t, "anAppName", applications[0].Name)
	assert.Equal(t, "anAppId", applications[0].Description)
}

func TestAzureClient_GetServicePrincipals_Errors(t *testing.T) {
	m := azuretestsupport.NewAzureHttpClient()
	tests := []struct {
		name        string
		badToken    bool
		badKey      bool
		badRequest  bool
		badResponse bool
		errSubstr   string
	}{
		{
			name:      "Bad Key",
			badKey:    true,
			errSubstr: "invalid character 'k'",
		},
		{
			name:      "Bad Access Token",
			badToken:  true,
			errSubstr: "invalid character 'a'",
		},
		{
			name:       "Unexpected Status",
			badRequest: true,
			errSubstr:  "Unexpected status",
		},
		{
			name:        "Bad Response",
			badResponse: true,
			errSubstr:   "invalid character '~'",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			accessToken := "accessToken"
			key := azuretestsupport.AzureClientKey()

			if tt.badKey {
				key = []byte("keyInvalid")
			}

			if tt.badToken {
				accessToken = `"accessToken"`
			}

			m.TokenRequest(accessToken)
			if tt.badRequest {
				m.ErrorRequest(http.MethodGet, m.GetServicePrincipalsUrl(), http.StatusBadRequest, nil)
			}

			if tt.badResponse {
				m.ErrorRequest(http.MethodGet, m.GetServicePrincipalsUrl(), http.StatusOK, []byte("~"))
			}

			client := m.AzureClient()
			sps, err := client.GetServicePrincipals(key, azuretestsupport.AzureAppId)
			assert.Error(t, err)
			assert.ErrorContains(t, err, tt.errSubstr)
			assert.NotNil(t, sps)
			assert.Empty(t, sps.List)
		})
	}
}

func TestAzureClient_GetServicePrincipals(t *testing.T) {
	m := azuretestsupport.NewAzureHttpClient()
	key := azuretestsupport.AzureClientKey()
	m.TokenRequest("accessToken")
	m.GetServicePrincipalsRequest()
	client := m.AzureClient()
	sps, err := client.GetServicePrincipals(key, azuretestsupport.AzureAppId)

	assert.NoError(t, err)
	assert.NotNil(t, sps)
	assert.Equal(t, 1, len(sps.List))
	assert.Equal(t, azuretestsupport.ServicePrincipalId, sps.List[0].ID)
	assert.Equal(t, policytestsupport.PolicyObjectResourceId, sps.List[0].Name)
	assert.NotNil(t, sps.List[0].AppRoles)

	expSps := azuretestsupport.AzureServicePrincipals()
	assert.Equal(t, len(expSps.List[0].AppRoles), len(sps.List[0].AppRoles))
}

func TestAzureClient_GetUserInfoFromPrincipalId_Errors(t *testing.T) {
	m := azuretestsupport.NewAzureHttpClient()
	principalId := policytestsupport.UserIdGetHrUsAndProfile
	getUrl := m.GetUserInfoFromPrincipalIdUrl(principalId)
	tests := []struct {
		name        string
		badToken    bool
		badKey      bool
		badRequest  bool
		badResponse bool
		errSubstr   string
	}{
		{
			name:      "Bad Key",
			badKey:    true,
			errSubstr: "invalid character 'k'",
		},
		{
			name:      "Bad Access Token",
			badToken:  true,
			errSubstr: "invalid character 'a'",
		},
		{
			name:       "Unexpected Status",
			badRequest: true,
			errSubstr:  "Unexpected status",
		},
		{
			name:        "Bad Response",
			badResponse: true,
			errSubstr:   "invalid character '~'",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			accessToken := "accessToken"
			key := azuretestsupport.AzureClientKey()

			if tt.badKey {
				key = []byte("keyInvalid")
			}

			if tt.badToken {
				accessToken = `"accessToken"`
			}

			m.TokenRequest(accessToken)
			if tt.badRequest {
				m.ErrorRequest(http.MethodGet, getUrl, http.StatusBadRequest, nil)
			}

			if tt.badResponse {
				m.ErrorRequest(http.MethodGet, getUrl, http.StatusOK, []byte("~"))
			}

			client := m.AzureClient()
			actUser, err := client.GetUserInfoFromPrincipalId(key, principalId)
			assert.Error(t, err)
			assert.ErrorContains(t, err, tt.errSubstr)
			assert.NotNil(t, actUser)
			assert.Empty(t, actUser)
		})
	}
}

func TestAzureClient_GetUserInfoFromPrincipalId(t *testing.T) {
	m := azuretestsupport.NewAzureHttpClient()
	key := azuretestsupport.AzureClientKey()
	principalId := policytestsupport.UserIdGetProfile
	m.TokenRequest("accessToken")
	m.GetUserInfoFromPrincipalIdRequest(principalId)

	client := m.AzureClient()
	actUser, err := client.GetUserInfoFromPrincipalId(key, principalId)
	assert.NoError(t, err)
	assert.NotNil(t, actUser)
	assert.Equal(t, principalId, actUser.PrincipalId)
	assert.Equal(t, policytestsupport.MakeEmail(principalId), actUser.Email)
}

func TestAzureClient_GetPrincipalIdFromEmail_Errors(t *testing.T) {
	m := azuretestsupport.NewAzureHttpClient()
	principalId := policytestsupport.UserIdGetHrUs
	email := policytestsupport.UserEmailGetHrUs
	getUrl := m.GetPrincipalIdFromEmailUrl(principalId)

	tests := []struct {
		name        string
		badToken    bool
		badKey      bool
		badRequest  bool
		badResponse bool
		errSubstr   string
	}{
		{
			name:      "Bad Key",
			badKey:    true,
			errSubstr: "invalid character 'k'",
		},
		{
			name:      "Bad Access Token",
			badToken:  true,
			errSubstr: "invalid character 'a'",
		},
		{
			name:       "Unexpected Status",
			badRequest: true,
			errSubstr:  "Unexpected status",
		},
		{
			name:        "Bad Response",
			badResponse: true,
			errSubstr:   "invalid character '~'",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			accessToken := "accessToken"
			key := azuretestsupport.AzureClientKey()

			if tt.badKey {
				key = []byte("keyInvalid")
			}

			if tt.badToken {
				accessToken = `"accessToken"`
			}

			m.TokenRequest(accessToken)
			if tt.badRequest {
				m.ErrorRequest(http.MethodGet, getUrl, http.StatusBadRequest, nil)
			}

			if tt.badResponse {
				m.ErrorRequest(http.MethodGet, getUrl, http.StatusOK, []byte("~"))
			}

			client := m.AzureClient()
			actPrincipalId, err := client.GetPrincipalIdFromEmail(key, email)
			assert.Error(t, err)
			assert.ErrorContains(t, err, tt.errSubstr)
			assert.Empty(t, actPrincipalId)
		})
	}
}

func TestAzureClient_GetPrincipalIdFromEmail(t *testing.T) {
	m := azuretestsupport.NewAzureHttpClient()
	key := azuretestsupport.AzureClientKey()
	m.TokenRequest("accessToken")
	principalId := policytestsupport.UserIdGetHrUs
	email := policytestsupport.UserEmailGetHrUs
	m.GetPrincipalIdFromEmailRequest(principalId)

	client := m.AzureClient()
	actPrincipalId, err := client.GetPrincipalIdFromEmail(key, email)
	assert.NoError(t, err)
	assert.Equal(t, principalId, actPrincipalId)
}

func TestAzureClient_GetAppRoleAssignedTo_Errors(t *testing.T) {
	m := azuretestsupport.NewAzureHttpClient()
	getUrl := m.AppRoleAssignmentsUrl()

	tests := []struct {
		name        string
		badToken    bool
		badKey      bool
		badRequest  bool
		badResponse bool
		errSubstr   string
	}{
		{
			name:      "Bad Key",
			badKey:    true,
			errSubstr: "invalid character 'k'",
		},
		{
			name:      "Bad Access Token",
			badToken:  true,
			errSubstr: "invalid character 'a'",
		},
		{
			name:       "Unexpected Status",
			badRequest: true,
			errSubstr:  "Unexpected status",
		},
		{
			name:        "Bad Response",
			badResponse: true,
			errSubstr:   "invalid character '~'",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			accessToken := "accessToken"
			key := azuretestsupport.AzureClientKey()

			if tt.badKey {
				key = []byte("keyInvalid")
			}

			if tt.badToken {
				accessToken = `"accessToken"`
			}

			m.TokenRequest(accessToken)
			if tt.badRequest {
				m.ErrorRequest(http.MethodGet, getUrl, http.StatusBadRequest, nil)
			}

			if tt.badResponse {
				m.ErrorRequest(http.MethodGet, getUrl, http.StatusOK, []byte("~"))
			}

			client := m.AzureClient()
			actAssignments, err := client.GetAppRoleAssignedTo(key, azuretestsupport.ServicePrincipalId)
			assert.Error(t, err)
			assert.ErrorContains(t, err, tt.errSubstr)
			assert.Empty(t, actAssignments)
		})
	}
}

func TestAzureClient_GetAppRoleAssignedTo_NoAssignments(t *testing.T) {
	m := azuretestsupport.NewAzureHttpClient()
	key := azuretestsupport.AzureClientKey()
	var existingAssignments []microsoftazure.AzureAppRoleAssignment

	m.TokenRequest("accessToken")
	m.GetAppRoleAssignmentsRequest(existingAssignments)

	client := m.AzureClient()
	actAssignments, err := client.GetAppRoleAssignedTo(key, azuretestsupport.ServicePrincipalId)
	assert.NoError(t, err)
	assert.NotNil(t, actAssignments.List)
	assert.Equal(t, 0, len(actAssignments.List))
}

func TestAzureClient_GetAppRoleAssignedTo(t *testing.T) {
	m := azuretestsupport.NewAzureHttpClient()
	key := azuretestsupport.AzureClientKey()
	existingAssignments := azuretestsupport.AppRoleAssignmentGetHrUs

	m.TokenRequest("accessToken")
	m.GetAppRoleAssignmentsRequest(existingAssignments)

	client := m.AzureClient()
	actAssignments, err := client.GetAppRoleAssignedTo(key, azuretestsupport.ServicePrincipalId)
	assert.NoError(t, err)
	assert.NotNil(t, actAssignments.List)
	assert.Equal(t, len(existingAssignments), len(actAssignments.List))

	existingAssignmentIdMap := make(map[string]microsoftazure.AzureAppRoleAssignment)
	for _, ara := range existingAssignments {
		existingAssignmentIdMap[ara.ID] = ara
	}

	foundCount := 0
	for _, ara := range actAssignments.List {
		exAra, found := existingAssignmentIdMap[ara.ID]
		if found {
			assert.Equal(t, exAra.AppRoleId, ara.AppRoleId)
			assert.Equal(t, exAra.PrincipalId, ara.PrincipalId)
			assert.Equal(t, exAra.ResourceId, ara.ResourceId)
			foundCount++
		}
	}

	assert.Equal(t, len(existingAssignments), foundCount)
}

func TestAzureClient_SetAppRoleAssignedTo_InvalidAppRoleAssignment(t *testing.T) {
	m := azuretestsupport.NewAzureHttpClient()
	client := m.AzureClient()

	tests := []struct {
		name       string
		assignment microsoftazure.AzureAppRoleAssignment
		errSubstr  string
	}{
		{
			name:       "Missing appRoleId",
			assignment: microsoftazure.AzureAppRoleAssignment{PrincipalId: "aUserId", ResourceId: "aResourceId"},
			errSubstr:  "AppRoleId",
		},
		{
			name:       "Missing Resource",
			assignment: microsoftazure.AzureAppRoleAssignment{AppRoleId: "aRoleId", PrincipalId: "aUserId"},
			errSubstr:  "ResourceId",
		},
	}

	key := []byte("xx")
	appId := "anAppId"
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := client.SetAppRoleAssignedTo(key, appId, []microsoftazure.AzureAppRoleAssignment{tt.assignment})
			assert.Error(t, err)
			assert.ErrorContains(t, err, "Error:Field validation")
			assert.ErrorContains(t, err, tt.errSubstr)
		})
	}
}

func TestAzureClient_SetAppRoleAssignedTo_WithBadGetAssignments(t *testing.T) {
	m := azuretestsupport.NewAzureHttpClient()
	m.TokenRequest("accessToken")
	m.ErrorRequest(http.MethodGet, m.AppRoleAssignmentsUrl(), http.StatusOK, []byte(`~`))
	client := m.AzureClient()
	err := client.SetAppRoleAssignedTo(
		azuretestsupport.AzureClientKey(),
		azuretestsupport.ServicePrincipalId,
		[]microsoftazure.AzureAppRoleAssignment{})
	assert.Error(t, err)
	assert.ErrorContains(t, err, "invalid character '~'")
}

func TestAzureClient_SetAppRoleAssignedTo_withBadAdd(t *testing.T) {
	m := azuretestsupport.NewAzureHttpClient()
	existingAssignments := azuretestsupport.AppRoleAssignmentGetHrUs

	m.TokenRequest("accessToken")
	m.GetAppRoleAssignmentsRequest(existingAssignments)
	m.ErrorRequest(http.MethodPost, m.AppRoleAssignmentsUrl(), http.StatusBadRequest, nil)

	newRoleAssignments := append(existingAssignments, azuretestsupport.AppRoleAssignmentForAdd...)

	client := m.AzureClient()
	err := client.SetAppRoleAssignedTo(
		azuretestsupport.AzureClientKey(),
		azuretestsupport.ServicePrincipalId,
		newRoleAssignments)
	assert.Error(t, err)
	assert.ErrorContains(t, err, "Unexpected status 400")
	assert.True(t, m.MockHttpClient.VerifyCalled())
}

func TestAzureClient_SetAppRoleAssignedTo_withBadDelete(t *testing.T) {
	m := azuretestsupport.NewAzureHttpClient()

	existingAssignments := azuretestsupport.AppRoleAssignmentGetHrUs

	m.TokenRequest("accessToken")
	m.GetAppRoleAssignmentsRequest(existingAssignments)
	deleteUrl := m.AppRoleAssignmentsUrl() + "/" + existingAssignments[0].ID
	m.ErrorRequest(http.MethodDelete, deleteUrl, http.StatusBadRequest, nil)

	assignmentsFromPolicy := azuretestsupport.AssignmentsForDelete(existingAssignments)
	client := m.AzureClient()
	err := client.SetAppRoleAssignedTo(
		azuretestsupport.AzureClientKey(),
		azuretestsupport.ServicePrincipalId,
		assignmentsFromPolicy)

	assert.Error(t, err)
	assert.ErrorContains(t, err, "Unexpected status 400")
	assert.True(t, m.MockHttpClient.VerifyCalled())
}

func TestAzureClient_SetAppRoleAssignedTo_AddNewRoleAssignment(t *testing.T) {
	m := azuretestsupport.NewAzureHttpClient()
	existingAssignments := azuretestsupport.AppRoleAssignmentGetHrUs

	m.TokenRequest("accessToken")
	m.GetAppRoleAssignmentsRequest(existingAssignments)
	m.PostAppRoleAssignmentsRequest()

	newRoleAssignments := make([]microsoftazure.AzureAppRoleAssignment, 0)
	newRoleAssignments = append(newRoleAssignments, azuretestsupport.AppRoleAssignmentForAdd...)
	newRoleAssignments = append(newRoleAssignments, existingAssignments...)

	client := m.AzureClient()
	err := client.SetAppRoleAssignedTo(
		azuretestsupport.AzureClientKey(),
		azuretestsupport.ServicePrincipalId,
		newRoleAssignments)
	assert.NoError(t, err)

	assert.True(t, m.MockHttpClient.VerifyCalled())
}

func TestAzureClient_SetAppRoleAssignedTo_AddOneMoreMember_ExistingRole(t *testing.T) {
	m := azuretestsupport.NewAzureHttpClient()
	existingAssignments := azuretestsupport.AppRoleAssignmentGetHrUs

	m.TokenRequest("accessToken")
	m.GetAppRoleAssignmentsRequest(existingAssignments)
	m.PostAppRoleAssignmentsRequest()

	newRoleAssignments := make([]microsoftazure.AzureAppRoleAssignment, 0)
	newRoleAssignments = append(newRoleAssignments, existingAssignments...)
	newRoleAssignments = append(newRoleAssignments, azuretestsupport.NewAppRoleAssignments(azuretestsupport.AppRoleIdGetHrUs, policytestsupport.UserIdUnassigned1))

	client := m.AzureClient()
	err := client.SetAppRoleAssignedTo(
		azuretestsupport.AzureClientKey(),
		azuretestsupport.ServicePrincipalId,
		newRoleAssignments)
	assert.NoError(t, err)

	assert.True(t, m.MockHttpClient.VerifyCalled())
}

func TestAzureClient_SetAppRoleAssignedTo_AddMoreMembers_ExistingRole(t *testing.T) {
	m := azuretestsupport.NewAzureHttpClient()
	existingAssignments := azuretestsupport.AppRoleAssignmentGetHrUs

	m.TokenRequest("accessToken")
	m.GetAppRoleAssignmentsRequest(existingAssignments)
	m.PostAppRoleAssignmentsRequest()

	newRoleAssignments := make([]microsoftazure.AzureAppRoleAssignment, 0)
	newRoleAssignments = append(newRoleAssignments, existingAssignments...)
	newRoleAssignments = append(newRoleAssignments, azuretestsupport.NewAppRoleAssignments(azuretestsupport.AppRoleIdGetHrUs, policytestsupport.UserIdUnassigned1))
	newRoleAssignments = append(newRoleAssignments, azuretestsupport.NewAppRoleAssignments(azuretestsupport.AppRoleIdGetHrUs, policytestsupport.UserIdUnassigned2))

	client := m.AzureClient()
	err := client.SetAppRoleAssignedTo(
		azuretestsupport.AzureClientKey(),
		azuretestsupport.ServicePrincipalId,
		newRoleAssignments)
	assert.NoError(t, err)

	assert.True(t, m.MockHttpClient.VerifyCalled())
}

func TestAzureClient_SetAppRoleAssignedTo_DoesNotAddExistingAssignment(t *testing.T) {
	m := azuretestsupport.NewAzureHttpClient()
	existingAssignments := azuretestsupport.AppRoleAssignmentGetHrUs

	m.TokenRequest("accessToken")
	m.GetAppRoleAssignmentsRequest(existingAssignments)

	client := m.AzureClient()
	err := client.SetAppRoleAssignedTo(
		azuretestsupport.AzureClientKey(),
		azuretestsupport.ServicePrincipalId,
		existingAssignments)
	assert.NoError(t, err)

	assert.True(t, m.MockHttpClient.VerifyCalled())
}

func TestAzureClient_SetAppRoleAssignedTo_DoesNotDeleteMemberInPolicy(t *testing.T) {
	m := azuretestsupport.NewAzureHttpClient()

	existingAssignments := azuretestsupport.AppRoleAssignmentGetHrUsAndProfile

	m.TokenRequest("accessToken")
	m.GetAppRoleAssignmentsRequest(existingAssignments)
	assignmentsFromPolicy := existingAssignments

	client := m.AzureClient()
	err := client.SetAppRoleAssignedTo(
		azuretestsupport.AzureClientKey(),
		azuretestsupport.ServicePrincipalId,
		assignmentsFromPolicy)

	assert.NoError(t, err)
	assert.True(t, m.MockHttpClient.VerifyCalled())
}

func TestAzureClient_SetAppRoleAssignedTo_DoesNotDeleteDifferentRole(t *testing.T) {
	m := azuretestsupport.NewAzureHttpClient()
	existingAssignments := azuretestsupport.AppRoleAssignmentGetHrUs
	assignmentsFromPolicy := azuretestsupport.AssignmentsForDelete(azuretestsupport.AppRoleAssignmentGetProfile)

	m.TokenRequest("accessToken")
	m.GetAppRoleAssignmentsRequest(existingAssignments)

	client := m.AzureClient()
	err := client.SetAppRoleAssignedTo(
		azuretestsupport.AzureClientKey(),
		azuretestsupport.ServicePrincipalId,
		assignmentsFromPolicy)

	assert.NoError(t, err)
	assert.True(t, m.MockHttpClient.VerifyCalled())
}

func TestAzureClient_SetAppRoleAssignedTo_DoesNotDeleteDifferentResource(t *testing.T) {
	m := azuretestsupport.NewAzureHttpClient()
	existingAssignments := azuretestsupport.AppRoleAssignmentGetHrUs
	assignmentsFromPolicy := make([]microsoftazure.AzureAppRoleAssignment, 0)

	m.TokenRequest("accessToken")
	m.GetAppRoleAssignmentsRequest(existingAssignments)

	for _, ara := range existingAssignments {
		newAra := microsoftazure.AzureAppRoleAssignment{
			AppRoleId:  ara.AppRoleId,
			ResourceId: "some-other-resource",
		}
		assignmentsFromPolicy = append(assignmentsFromPolicy, newAra)
	}

	client := m.AzureClient()
	err := client.SetAppRoleAssignedTo(
		azuretestsupport.AzureClientKey(),
		azuretestsupport.ServicePrincipalId,
		assignmentsFromPolicy)

	assert.NoError(t, err)
	assert.True(t, m.MockHttpClient.VerifyCalled())
}

func TestAzureClient_SetAppRoleAssignedTo_DeleteAllRoleAssignments(t *testing.T) {
	m := azuretestsupport.NewAzureHttpClient()
	existingAssignments := azuretestsupport.AppRoleAssignmentGetHrUsAndProfile
	assignmentsFromPolicy := azuretestsupport.AssignmentsForDelete(existingAssignments)

	m.TokenRequest("accessToken")
	m.GetAppRoleAssignmentsRequest(existingAssignments)
	m.DeleteAppRoleAssignmentsRequest(existingAssignments)

	client := m.AzureClient()
	err := client.SetAppRoleAssignedTo(
		azuretestsupport.AzureClientKey(),
		azuretestsupport.ServicePrincipalId,
		assignmentsFromPolicy)

	assert.NoError(t, err)
	assert.True(t, m.MockHttpClient.VerifyCalled())
}

func TestAzureClient_SetAppRoleAssignedTo_DeleteOneOfMultipleMemberAssignments(t *testing.T) {
	m := azuretestsupport.NewAzureHttpClient()
	existingAssignments := azuretestsupport.AppRoleAssignmentMultipleMembers
	assignmentsFromPolicy := []microsoftazure.AzureAppRoleAssignment{
		existingAssignments[1],
	}

	m.TokenRequest("accessToken")
	m.GetAppRoleAssignmentsRequest(existingAssignments)
	m.DeleteAppRoleAssignmentsRequest(existingAssignments[0:1])

	client := m.AzureClient()
	err := client.SetAppRoleAssignedTo(
		azuretestsupport.AzureClientKey(),
		azuretestsupport.ServicePrincipalId,
		assignmentsFromPolicy)

	assert.NoError(t, err)
	assert.True(t, m.MockHttpClient.VerifyCalled())
}

func TestAzureClient_SetAppRoleAssignedTo_AddDeleteRoleAssignment(t *testing.T) {
	m := azuretestsupport.NewAzureHttpClient()
	existingAssignments := azuretestsupport.AppRoleAssignmentGetHrUs
	newRoleAssignments := azuretestsupport.AppRoleAssignmentForAdd
	assignmentsFromPolicy := make([]microsoftazure.AzureAppRoleAssignment, 0)
	assignmentsFromPolicy = append(assignmentsFromPolicy, newRoleAssignments...)
	assignmentsFromPolicy = append(assignmentsFromPolicy, azuretestsupport.AssignmentsForDelete(existingAssignments)...)

	m.TokenRequest("accessToken")
	m.GetAppRoleAssignmentsRequest(existingAssignments)
	m.PostAppRoleAssignmentsRequest()
	m.DeleteAppRoleAssignmentsRequest(existingAssignments)

	client := m.AzureClient()
	err := client.SetAppRoleAssignedTo(
		azuretestsupport.AzureClientKey(),
		azuretestsupport.ServicePrincipalId,
		assignmentsFromPolicy)
	assert.NoError(t, err)
	assert.True(t, m.MockHttpClient.VerifyCalled())
}
