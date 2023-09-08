package microsoftazure_test

import (
	"github.com/hexa-org/policy-orchestrator/internal/orchestrator"
	"github.com/hexa-org/policy-orchestrator/internal/orchestratorproviders/microsoftazure"
	"github.com/hexa-org/policy-orchestrator/internal/orchestratorproviders/microsoftazure/azad"
	"github.com/hexa-org/policy-orchestrator/pkg/azuretestsupport"
	"github.com/hexa-org/policy-orchestrator/pkg/hexapolicy"
	"github.com/hexa-org/policy-orchestrator/pkg/testsupport/policytestsupport"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"log"
	"net/http"
	"testing"
)

func TestDiscoverApplications(t *testing.T) {
	key := azuretestsupport.AzureKeyBytes()
	mockAzClient := azuretestsupport.NewMockAzureClient()
	expApps := []orchestrator.ApplicationInfo{
		{
			ObjectID:    "anId",
			Name:        "aName",
			Description: "aDescription",
			Service:     "App Service",
		},
	}
	mockAzClient.On("GetWebApplications", key).Return(expApps, nil)

	p := microsoftazure.NewAzureProvider(microsoftazure.WithAzureClient(mockAzClient))

	info := orchestrator.IntegrationInfo{Name: "azure", Key: key}
	applications, _ := p.DiscoverApplications(info)
	log.Println(applications[0])

	assert.Len(t, applications, 1)
	assert.Equal(t, "azure", p.Name())
	assert.Equal(t, "App Service", applications[0].Service)
	mockAzClient.AssertExpectations(t)
}

func TestGetPolicy_WithoutUserEmail(t *testing.T) {
	appId := azuretestsupport.AzureAppId
	key := azuretestsupport.AzureKeyBytes()

	mockAzClient := azuretestsupport.NewMockAzureClient()
	mockAzClient.ExpectGetServicePrincipals()
	mockAzClient.ExpectAppRoleAssignedTo(azuretestsupport.AppRoleAssignmentGetProfile)
	mockAzClient.On("GetUserInfoFromPrincipalId", mock.Anything, mock.Anything).
		Return(azad.AzureUser{
			PrincipalId: policytestsupport.UserIdGetProfile,
		}, nil)

	p := microsoftazure.NewAzureProvider(microsoftazure.WithAzureClient(mockAzClient))

	info := orchestrator.IntegrationInfo{Name: "azure", Key: key}
	appInfo := orchestrator.ApplicationInfo{ObjectID: "anObjectId", Name: "anAppName", Description: appId}

	actualPolicies, err := p.GetPolicyInfo(info, appInfo)
	assert.NoError(t, err)
	assert.NotNil(t, actualPolicies)
	assert.Equal(t, len(azuretestsupport.AzureServicePrincipals().List[0].AppRoles), len(actualPolicies))

	for _, pol := range actualPolicies {
		assert.True(t, len(pol.Actions) > 0)
		assert.NotEmpty(t, pol.Actions[0].ActionUri)
		assert.Equal(t, 0, len(pol.Subject.Members))
		assert.Equal(t, policytestsupport.PolicyObjectResourceId, pol.Object.ResourceID)
	}
	mockAzClient.AssertExpectations(t)
}

func TestGetPolicy_WithRoleAssignment(t *testing.T) {
	appId := azuretestsupport.AzureAppId
	key := azuretestsupport.AzureKeyBytes()
	expAssignments := azuretestsupport.AppRoleAssignmentGetHrUs

	mockAzClient := azuretestsupport.NewMockAzureClient()
	mockAzClient.ExpectGetServicePrincipals()
	mockAzClient.ExpectAppRoleAssignedTo(expAssignments)
	mockAzClient.ExpectGetUserInfoFromPrincipalId(policytestsupport.UserIdGetHrUs)

	p := microsoftazure.NewAzureProvider(microsoftazure.WithAzureClient(mockAzClient))

	info := orchestrator.IntegrationInfo{Name: "azure", Key: key}
	appInfo := orchestrator.ApplicationInfo{ObjectID: "anObjectId", Name: "anAppName", Description: appId}

	actualPolicies, err := p.GetPolicyInfo(info, appInfo)
	assert.NoError(t, err)
	assert.NotNil(t, actualPolicies)
	assert.Equal(t, len(azuretestsupport.AzureServicePrincipals().List[0].AppRoles), len(actualPolicies))

	expPolicies := azuretestsupport.MakePolicies(expAssignments)
	assert.Equal(t, len(expPolicies), len(actualPolicies))
	assert.True(t, policytestsupport.ContainsPolicies(t, expPolicies, actualPolicies))
	mockAzClient.AssertExpectations(t)
}

func TestGetPolicy_MultiplePolicies(t *testing.T) {
	appId := azuretestsupport.AzureAppId
	key := azuretestsupport.AzureKeyBytes()
	expAssignments := azuretestsupport.AppRoleAssignmentGetHrUsAndProfile

	mockAzClient := azuretestsupport.NewMockAzureClient()
	mockAzClient.ExpectGetServicePrincipals()
	mockAzClient.ExpectAppRoleAssignedTo(expAssignments)
	mockAzClient.ExpectGetUserInfoFromPrincipalId(policytestsupport.UserIdGetHrUsAndProfile)

	p := microsoftazure.NewAzureProvider(microsoftazure.WithAzureClient(mockAzClient))

	info := orchestrator.IntegrationInfo{Name: "azure", Key: key}
	appInfo := orchestrator.ApplicationInfo{ObjectID: "anObjectId", Name: "anAppName", Description: appId}

	actualPolicies, err := p.GetPolicyInfo(info, appInfo)
	assert.NoError(t, err)
	assert.NotNil(t, actualPolicies)
	assert.Equal(t, len(azuretestsupport.AzureServicePrincipals().List[0].AppRoles), len(actualPolicies))

	expPolicies := azuretestsupport.MakePolicies(expAssignments)
	assert.Equal(t, len(expPolicies), len(actualPolicies))
	assert.True(t, policytestsupport.ContainsPolicies(t, expPolicies, actualPolicies))
	mockAzClient.AssertExpectations(t)
}

func TestGetPolicy_MultipleMembersInOnePolicy(t *testing.T) {
	appId := azuretestsupport.AzureAppId
	key := azuretestsupport.AzureKeyBytes()
	expAssignments := azuretestsupport.AppRoleAssignmentMultipleMembers

	mockAzClient := azuretestsupport.NewMockAzureClient()
	mockAzClient.ExpectGetServicePrincipals()
	mockAzClient.ExpectAppRoleAssignedTo(expAssignments)
	mockAzClient.ExpectGetUserInfoFromPrincipalId(policytestsupport.UserIdGetHrUs, policytestsupport.UserIdGetHrUsAndProfile)

	p := microsoftazure.NewAzureProvider(microsoftazure.WithAzureClient(mockAzClient))

	info := orchestrator.IntegrationInfo{Name: "azure", Key: key}
	appInfo := orchestrator.ApplicationInfo{ObjectID: "anObjectId", Name: "anAppName", Description: appId}

	actualPolicies, err := p.GetPolicyInfo(info, appInfo)
	assert.NoError(t, err)
	assert.NotNil(t, actualPolicies)
	assert.Equal(t, len(azuretestsupport.AzureServicePrincipals().List[0].AppRoles), len(actualPolicies))

	expPolicies := azuretestsupport.MakePolicies(expAssignments)
	assert.Equal(t, len(expPolicies), len(actualPolicies))
	assert.True(t, policytestsupport.ContainsPolicies(t, expPolicies, actualPolicies))
	mockAzClient.AssertExpectations(t)
}

func TestSetPolicy_withInvalidArguments(t *testing.T) {
	azureProvider := microsoftazure.NewAzureProvider()
	key := []byte("key")

	status, err := azureProvider.SetPolicyInfo(
		orchestrator.IntegrationInfo{Name: "azure", Key: key},
		orchestrator.ApplicationInfo{Name: "anAppName", Description: "anAppId"},
		[]hexapolicy.PolicyInfo{{
			Meta:    hexapolicy.MetaInfo{Version: "0"},
			Actions: []hexapolicy.ActionInfo{{"azure:anAppRoleId"}},
			Subject: hexapolicy.SubjectInfo{Members: []string{"aPrincipalId:aPrincipalDisplayName", "yetAnotherPrincipalId:yetAnotherPrincipalDisplayName", "andAnotherPrincipalId:andAnotherPrincipalDisplayName"}},
			Object: hexapolicy.ObjectInfo{
				ResourceID: "anObjectId",
			},
		}})

	assert.Equal(t, http.StatusInternalServerError, status)
	assert.EqualError(t, err, "Key: 'ApplicationInfo.ObjectID' Error:Field validation for 'ObjectID' failed on the 'required' tag")

	status, err = azureProvider.SetPolicyInfo(
		orchestrator.IntegrationInfo{Name: "azure", Key: key},
		orchestrator.ApplicationInfo{ObjectID: "anObjectId", Name: "anAppName", Description: "aDescription"},
		[]hexapolicy.PolicyInfo{{
			Meta:    hexapolicy.MetaInfo{Version: "0"},
			Actions: []hexapolicy.ActionInfo{{"azure:anAppRoleId"}},
			Subject: hexapolicy.SubjectInfo{Members: []string{"aPrincipalId:aPrincipalDisplayName", "yetAnotherPrincipalId:yetAnotherPrincipalDisplayName", "andAnotherPrincipalId:andAnotherPrincipalDisplayName"}},
			Object:  hexapolicy.ObjectInfo{},
		}})

	assert.Equal(t, http.StatusInternalServerError, status)
	assert.EqualError(t, err, "Key: '[0].Object.ResourceID' Error:Field validation for 'ResourceID' failed on the 'required' tag")
}

func TestSetPolicy_IgnoresAllPrincipalIdsNotFound(t *testing.T) {
	appId := azuretestsupport.AzureAppId
	key := azuretestsupport.AzureKeyBytes()

	mockAzClient := azuretestsupport.NewMockAzureClient()
	mockAzClient.ExpectGetServicePrincipals()

	mockAzClient.ExpectGetPrincipalIdFromEmail(policytestsupport.UserEmailGetHrUs, "")
	mockAzClient.ExpectGetPrincipalIdFromEmail(policytestsupport.UserEmailGetProfile, "")

	p := microsoftazure.NewAzureProvider(microsoftazure.WithAzureClient(mockAzClient))
	status, err := p.SetPolicyInfo(
		orchestrator.IntegrationInfo{Name: "azure", Key: key},
		orchestrator.ApplicationInfo{ObjectID: "anObjectId", Name: "anAppName", Description: appId},
		[]hexapolicy.PolicyInfo{{
			Meta:    hexapolicy.MetaInfo{Version: "0"},
			Actions: []hexapolicy.ActionInfo{{"azure:" + policytestsupport.ActionGetHrUs}},
			Subject: hexapolicy.SubjectInfo{Members: []string{"user:" + policytestsupport.UserEmailGetHrUs,
				"user:" + policytestsupport.UserEmailGetProfile}},
			Object: hexapolicy.ObjectInfo{
				ResourceID: policytestsupport.PolicyObjectResourceId,
			},
		}})

	assert.NoError(t, err)
	assert.Equal(t, http.StatusCreated, status)
	mockAzClient.AssertExpectations(t)
}

func TestSetPolicy_IgnoresAnyNotFoundPrincipalId(t *testing.T) {
	appId := azuretestsupport.AzureAppId
	key := azuretestsupport.AzureKeyBytes()

	mockAzClient := azuretestsupport.NewMockAzureClient()
	mockAzClient.ExpectGetServicePrincipals()

	mockAzClient.ExpectGetPrincipalIdFromEmail(policytestsupport.UserEmailGetHrUs, policytestsupport.UserIdGetHrUs)
	mockAzClient.ExpectGetPrincipalIdFromEmail(policytestsupport.UserEmailGetProfile, "")
	mockAzClient.ExpectSetAppRoleAssignedTo(azuretestsupport.AppRoleAssignmentGetHrUs)

	p := microsoftazure.NewAzureProvider(microsoftazure.WithAzureClient(mockAzClient))
	status, err := p.SetPolicyInfo(
		orchestrator.IntegrationInfo{Name: "azure", Key: key},
		orchestrator.ApplicationInfo{ObjectID: "anObjectId", Name: "anAppName", Description: appId},
		[]hexapolicy.PolicyInfo{{
			Meta:    hexapolicy.MetaInfo{Version: "0"},
			Actions: []hexapolicy.ActionInfo{{"azure:" + policytestsupport.ActionGetHrUs}},
			Subject: hexapolicy.SubjectInfo{Members: []string{"user:" + policytestsupport.UserEmailGetHrUs,
				"user:" + policytestsupport.UserEmailGetProfile}},
			Object: hexapolicy.ObjectInfo{
				ResourceID: policytestsupport.PolicyObjectResourceId,
			},
		}})

	assert.NoError(t, err)
	assert.Equal(t, http.StatusCreated, status)
	mockAzClient.AssertExpectations(t)
}

func TestSetPolicy_AddAssignment_IgnoresInvalidAction(t *testing.T) {
	appId := azuretestsupport.AzureAppId
	key := azuretestsupport.AzureKeyBytes()

	mockAzClient := azuretestsupport.NewMockAzureClient()
	mockAzClient.ExpectGetServicePrincipals()

	p := microsoftazure.NewAzureProvider(microsoftazure.WithAzureClient(mockAzClient))
	status, err := p.SetPolicyInfo(
		orchestrator.IntegrationInfo{Name: "azure", Key: key},
		orchestrator.ApplicationInfo{ObjectID: "anObjectId", Name: "anAppName", Description: appId},
		[]hexapolicy.PolicyInfo{{
			Meta:    hexapolicy.MetaInfo{Version: "0"},
			Actions: []hexapolicy.ActionInfo{{"azure:GET/not_defined"}},
			Subject: hexapolicy.SubjectInfo{
				Members: []string{
					"user:" + policytestsupport.UserEmailGetHrUs,
					"user:" + policytestsupport.UserEmailGetProfile}},
			Object: hexapolicy.ObjectInfo{
				ResourceID: policytestsupport.PolicyObjectResourceId,
			},
		}})

	assert.NoError(t, err)
	assert.Equal(t, http.StatusCreated, status)
	mockAzClient.AssertExpectations(t)
}

func TestSetPolicy(t *testing.T) {
	appId := azuretestsupport.AzureAppId
	key := azuretestsupport.AzureKeyBytes()

	mockAzClient := azuretestsupport.NewMockAzureClient()
	mockAzClient.ExpectGetServicePrincipals()
	mockAzClient.ExpectGetPrincipalIdFromEmail(policytestsupport.UserEmailGetHrUs, policytestsupport.UserIdGetHrUs)
	mockAzClient.ExpectSetAppRoleAssignedTo(azuretestsupport.AppRoleAssignmentGetHrUs)

	p := microsoftazure.NewAzureProvider(microsoftazure.WithAzureClient(mockAzClient))
	status, err := p.SetPolicyInfo(
		orchestrator.IntegrationInfo{Name: "azure", Key: key},
		orchestrator.ApplicationInfo{ObjectID: "anObjectId", Name: "anAppName", Description: appId},
		[]hexapolicy.PolicyInfo{{
			Meta:    hexapolicy.MetaInfo{Version: "0"},
			Actions: []hexapolicy.ActionInfo{{"azure:" + policytestsupport.ActionGetHrUs}},
			Subject: hexapolicy.SubjectInfo{Members: []string{"user:" + policytestsupport.UserEmailGetHrUs}},
			Object: hexapolicy.ObjectInfo{
				ResourceID: policytestsupport.PolicyObjectResourceId,
			},
		}})

	assert.NoError(t, err)
	assert.Equal(t, http.StatusCreated, status)
	mockAzClient.AssertExpectations(t)
}

func TestSetPolicy_RemovedAllMembers_FromOnePolicy(t *testing.T) {
	appId := azuretestsupport.AzureAppId
	key := azuretestsupport.AzureKeyBytes()

	mockAzClient := azuretestsupport.NewMockAzureClient()
	mockAzClient.ExpectGetServicePrincipals()
	mockAzClient.ExpectSetAppRoleAssignedTo(
		azuretestsupport.AssignmentsForDelete(azuretestsupport.AppRoleAssignmentGetHrUs))

	p := microsoftazure.NewAzureProvider(microsoftazure.WithAzureClient(mockAzClient))
	status, err := p.SetPolicyInfo(
		orchestrator.IntegrationInfo{Name: "azure", Key: key},
		orchestrator.ApplicationInfo{ObjectID: "anObjectId", Name: "anAppName", Description: appId},
		[]hexapolicy.PolicyInfo{{
			Meta:    hexapolicy.MetaInfo{Version: "0"},
			Actions: []hexapolicy.ActionInfo{{"azure:" + policytestsupport.ActionGetHrUs}},
			Subject: hexapolicy.SubjectInfo{Members: []string{}},
			Object: hexapolicy.ObjectInfo{
				ResourceID: policytestsupport.PolicyObjectResourceId,
			},
		}})

	assert.NoError(t, err)
	assert.Equal(t, http.StatusCreated, status)
	mockAzClient.AssertExpectations(t)
}

func TestSetPolicy_RemovedAllMembers_FromAllPolicies(t *testing.T) {
	appId := azuretestsupport.AzureAppId
	key := azuretestsupport.AzureKeyBytes()

	mockAzClient := azuretestsupport.NewMockAzureClient()
	mockAzClient.ExpectGetServicePrincipals()
	mockAzClient.ExpectSetAppRoleAssignedTo(
		azuretestsupport.AssignmentsForDelete(azuretestsupport.AppRoleAssignmentGetHrUs))
	mockAzClient.ExpectSetAppRoleAssignedTo(
		azuretestsupport.AssignmentsForDelete(azuretestsupport.AppRoleAssignmentGetProfile))

	p := microsoftazure.NewAzureProvider(microsoftazure.WithAzureClient(mockAzClient))
	status, err := p.SetPolicyInfo(
		orchestrator.IntegrationInfo{Name: "azure", Key: key},
		orchestrator.ApplicationInfo{ObjectID: "anObjectId", Name: "anAppName", Description: appId},
		[]hexapolicy.PolicyInfo{
			{
				Meta:    hexapolicy.MetaInfo{Version: "0"},
				Actions: []hexapolicy.ActionInfo{{"azure:" + policytestsupport.ActionGetHrUs}},
				Subject: hexapolicy.SubjectInfo{Members: []string{}},
				Object: hexapolicy.ObjectInfo{
					ResourceID: policytestsupport.PolicyObjectResourceId,
				},
			},
			{
				Meta:    hexapolicy.MetaInfo{Version: "0"},
				Actions: []hexapolicy.ActionInfo{{"azure:" + policytestsupport.ActionGetProfile}},
				Subject: hexapolicy.SubjectInfo{Members: []string{}},
				Object: hexapolicy.ObjectInfo{
					ResourceID: policytestsupport.PolicyObjectResourceId,
				},
			},
		})

	assert.NoError(t, err)
	assert.Equal(t, http.StatusCreated, status)
	mockAzClient.AssertExpectations(t)
}

func TestSetPolicy_MultipleAppRolePolicies(t *testing.T) {
	appId := azuretestsupport.AzureAppId
	key := azuretestsupport.AzureKeyBytes()

	mockAzClient := azuretestsupport.NewMockAzureClient()
	mockAzClient.ExpectGetServicePrincipals()
	mockAzClient.ExpectGetPrincipalIdFromEmail(policytestsupport.UserEmailGetHrUs, policytestsupport.UserIdGetHrUs)
	mockAzClient.ExpectGetPrincipalIdFromEmail(policytestsupport.UserEmailGetProfile, policytestsupport.UserIdGetProfile)

	mockAzClient.ExpectSetAppRoleAssignedTo(azuretestsupport.AppRoleAssignmentGetHrUs)
	mockAzClient.ExpectSetAppRoleAssignedTo(azuretestsupport.AppRoleAssignmentGetProfile)

	p := microsoftazure.NewAzureProvider(microsoftazure.WithAzureClient(mockAzClient))
	status, err := p.SetPolicyInfo(
		orchestrator.IntegrationInfo{Name: "azure", Key: key},
		orchestrator.ApplicationInfo{ObjectID: "anObjectId", Name: "anAppName", Description: appId},
		[]hexapolicy.PolicyInfo{
			{
				Meta:    hexapolicy.MetaInfo{Version: "0"},
				Actions: []hexapolicy.ActionInfo{{"azure:" + policytestsupport.ActionGetHrUs}},
				Subject: hexapolicy.SubjectInfo{Members: []string{"user:" + policytestsupport.UserEmailGetHrUs}},
				Object: hexapolicy.ObjectInfo{
					ResourceID: policytestsupport.PolicyObjectResourceId,
				},
			},
			{
				Meta:    hexapolicy.MetaInfo{Version: "0"},
				Actions: []hexapolicy.ActionInfo{{"azure:" + policytestsupport.ActionGetProfile}},
				Subject: hexapolicy.SubjectInfo{Members: []string{"user:" + policytestsupport.UserEmailGetProfile}},
				Object: hexapolicy.ObjectInfo{
					ResourceID: policytestsupport.PolicyObjectResourceId,
				},
			},
		})

	assert.NoError(t, err)
	assert.Equal(t, http.StatusCreated, status)
	mockAzClient.AssertExpectations(t)
}
