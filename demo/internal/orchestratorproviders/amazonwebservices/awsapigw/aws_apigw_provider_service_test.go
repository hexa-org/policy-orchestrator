package awsapigw_test

import (
	"errors"
	"github.com/hexa-org/policy-mapper/pkg/hexapolicy"
	"github.com/hexa-org/policy-orchestrator/demo/internal/orchestrator"
	"github.com/hexa-org/policy-orchestrator/demo/internal/orchestratorproviders/amazonwebservices/awsapigw"
	"github.com/hexa-org/policy-orchestrator/demo/internal/orchestratorproviders/providerscommon"
	"github.com/hexa-org/policy-orchestrator/demo/pkg/testsupport/awstestsupport"
	"github.com/hexa-org/policy-orchestrator/demo/pkg/testsupport/policytestsupport"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"net/http"
	"testing"
)

type mockCognitoClient struct {
	mock.Mock
}

func (m *mockCognitoClient) ListUserPools() (apps []orchestrator.ApplicationInfo, err error) {
	args := m.Called()
	return args.Get(0).([]orchestrator.ApplicationInfo), args.Error(1)
}

func (m *mockCognitoClient) GetGroups(_ string) (map[string]string, error) {
	panic("GetGroups not implemented")
}

func (m *mockCognitoClient) GetMembersAssignedTo(_ orchestrator.ApplicationInfo, _ string) ([]string, error) {
	panic("GetMembersAssignedTo not implemented")
}

func (m *mockCognitoClient) SetGroupsAssignedTo(_ string, _ []string, _ orchestrator.ApplicationInfo) error {
	panic("SetGroupsAssignedTo not implemented")
}

func (m *mockCognitoClient) expectListUserPools(apps []orchestrator.ApplicationInfo, err error) {
	m.On("ListUserPools").Return(apps, err)
}

func TestAwsApiGatewayProviderService_DiscoverApplications_Error(t *testing.T) {
	cognitoClient := &mockCognitoClient{}
	cognitoClient.expectListUserPools(nil, errors.New("some error"))
	service := awsapigw.NewAwsApiGatewayProviderService(cognitoClient, nil)
	apps, err := service.DiscoverApplications(awstestsupport.IntegrationInfo())
	assert.Error(t, err)
	assert.Len(t, apps, 0)
}

func TestAwsApiGatewayProviderService_DiscoverApplications(t *testing.T) {
	cognitoClient := &mockCognitoClient{}
	expApps := []orchestrator.ApplicationInfo{awstestsupport.AppInfo()}
	cognitoClient.expectListUserPools(expApps, nil)
	service := awsapigw.NewAwsApiGatewayProviderService(cognitoClient, nil)
	apps, err := service.DiscoverApplications(awstestsupport.IntegrationInfo())
	assert.NoError(t, err)
	assert.Len(t, apps, len(expApps))
}

func TestGetPolicyInfo_Error(t *testing.T) {
	policyStoreSvc := &mockPolicyStoreSvc{}

	policyStoreSvc.expectGetResourceRoles(nil, errors.New("some-error"))
	service := awsapigw.NewAwsApiGatewayProviderService(nil, policyStoreSvc)

	appInfo := orchestrator.ApplicationInfo{}
	actPolicies, err := service.GetPolicyInfo(appInfo)
	assert.ErrorContains(t, err, "some-error")
	assert.NotNil(t, actPolicies)
	assert.Empty(t, actPolicies)
}

func TestGetPolicyInfo_NoResourceRoles(t *testing.T) {
	policyStoreSvc := &mockPolicyStoreSvc{}

	policyStoreSvc.expectGetResourceRoles(nil, nil)
	service := awsapigw.NewAwsApiGatewayProviderService(nil, policyStoreSvc)

	appInfo := orchestrator.ApplicationInfo{}
	actPolicies, err := service.GetPolicyInfo(appInfo)
	assert.NoError(t, err)
	assert.NotNil(t, actPolicies)
	assert.Empty(t, actPolicies)
}

func TestGetPolicyInfo(t *testing.T) {
	policyStoreSvc := &mockPolicyStoreSvc{}
	existingActionRoles := map[string][]string{
		policytestsupport.ActionGetHrUs:    {"some-hr-role"},
		policytestsupport.ActionGetProfile: {"some-profile-role"},
	}
	expReturnResourceRoles := policytestsupport.MakeRarList(existingActionRoles)

	policyStoreSvc.expectGetResourceRoles(expReturnResourceRoles, nil)
	service := awsapigw.NewAwsApiGatewayProviderService(nil, policyStoreSvc)

	appInfo := orchestrator.ApplicationInfo{}
	actPolicies, err := service.GetPolicyInfo(appInfo)
	assert.NoError(t, err)
	assert.NotNil(t, actPolicies)
	assert.Equal(t, len(existingActionRoles), len(actPolicies))

	for _, actPol := range actPolicies {
		actMembers := actPol.Subject.Members
		actPolRar := providerscommon.NewResourceActionUriRoles(actPol.Object.ResourceID, actPol.Actions[0].ActionUri, actMembers)
		actLookupKey := actPolRar.Action + actPolRar.Resource
		expMembers, found := existingActionRoles[actLookupKey]
		assert.True(t, found)
		assert.Equal(t, expMembers, actMembers)
	}
}

func TestSetPolicyInfo_GetResourcesError(t *testing.T) {
	policyStoreSvc := &mockPolicyStoreSvc{}
	policyStoreSvc.expectGetResourceRoles(nil, errors.New("some-error"))
	service := awsapigw.NewAwsApiGatewayProviderService(nil, policyStoreSvc)
	appInfo := orchestrator.ApplicationInfo{}
	status, err := service.SetPolicyInfo(appInfo, []hexapolicy.PolicyInfo{})
	assert.ErrorContains(t, err, "some-error")
	assert.Equal(t, http.StatusBadGateway, status)
}

func TestSetPolicyInfo_UpdateError(t *testing.T) {
	policyStoreSvc := &mockPolicyStoreSvc{}
	existingActionRoles := map[string][]string{
		policytestsupport.ActionGetHrUs:    {"some-hr-role"},
		policytestsupport.ActionGetProfile: {"some-profile-role"},
	}
	expReturnResourceRoles := policytestsupport.MakeRarList(existingActionRoles)
	policyStoreSvc.expectGetResourceRoles(expReturnResourceRoles, nil)

	newActionRoles := map[string][]string{
		policytestsupport.ActionGetHrUs:    {"some-hr-role-new"},
		policytestsupport.ActionGetProfile: {"some-profile-role-new"},
	}
	newResourceRoles := policytestsupport.MakeRarList(newActionRoles)
	policyStoreSvc.expectUpdateResourceRoles(newResourceRoles[0], errors.New("some-error"))

	policies := policytestsupport.MakeRoleSubjectTestPolicies(newActionRoles)

	service := awsapigw.NewAwsApiGatewayProviderService(nil, policyStoreSvc)
	appInfo := orchestrator.ApplicationInfo{}
	status, err := service.SetPolicyInfo(appInfo, policies)
	assert.ErrorContains(t, err, "some-error")
	assert.Equal(t, http.StatusBadGateway, status)
}

func TestSetPolicyInfo_MultiplePolicies(t *testing.T) {
	policyStoreSvc := &mockPolicyStoreSvc{}
	existingActionRoles := map[string][]string{
		policytestsupport.ActionGetHrUs:    {"some-hr-role"},
		policytestsupport.ActionGetProfile: {"some-profile-role"},
	}
	expReturnResourceRoles := policytestsupport.MakeRarList(existingActionRoles)
	policyStoreSvc.expectGetResourceRoles(expReturnResourceRoles, nil)

	newActionRoles := map[string][]string{
		policytestsupport.ActionGetHrUs:    {"some-hr-role-new"},
		policytestsupport.ActionGetProfile: {"some-profile-role-new"},
	}
	newResourceRoles := policytestsupport.MakeRarList(newActionRoles)
	policyStoreSvc.expectUpdateResourceRoles(newResourceRoles[0], nil)
	policyStoreSvc.expectUpdateResourceRoles(newResourceRoles[1], nil)

	policies := policytestsupport.MakeRoleSubjectTestPolicies(newActionRoles)

	service := awsapigw.NewAwsApiGatewayProviderService(nil, policyStoreSvc)
	appInfo := orchestrator.ApplicationInfo{}
	status, err := service.SetPolicyInfo(appInfo, policies)
	assert.NoError(t, err)
	assert.Equal(t, http.StatusCreated, status)
}

func TestSetPolicyInfo_UpdateInputPolicyOnly(t *testing.T) {
	policyStoreSvc := &mockPolicyStoreSvc{}
	existingActionRoles := map[string][]string{
		policytestsupport.ActionGetHrUs:    {"some-hr-role"},
		policytestsupport.ActionGetProfile: {"some-profile-role"},
	}
	expReturnResourceRoles := policytestsupport.MakeRarList(existingActionRoles)
	policyStoreSvc.expectGetResourceRoles(expReturnResourceRoles, nil)

	newActionRoles := map[string][]string{
		policytestsupport.ActionGetProfile: {"some-profile-role-new"},
	}
	newResourceRoles := policytestsupport.MakeRarList(newActionRoles)
	policyStoreSvc.expectUpdateResourceRoles(newResourceRoles[0], nil)

	policies := policytestsupport.MakeRoleSubjectTestPolicies(newActionRoles)

	service := awsapigw.NewAwsApiGatewayProviderService(nil, policyStoreSvc)
	appInfo := orchestrator.ApplicationInfo{}
	status, err := service.SetPolicyInfo(appInfo, policies)
	assert.NoError(t, err)
	assert.Equal(t, http.StatusCreated, status)
}

func TestSetPolicyInfo_RemoveAllMembers(t *testing.T) {
	policyStoreSvc := &mockPolicyStoreSvc{}
	existingActionRoles := map[string][]string{
		policytestsupport.ActionGetHrUs:    {"some-hr-role"},
		policytestsupport.ActionGetProfile: {"some-profile-role"},
	}
	expReturnResourceRoles := policytestsupport.MakeRarList(existingActionRoles)
	policyStoreSvc.expectGetResourceRoles(expReturnResourceRoles, nil)

	newActionRoles := map[string][]string{
		policytestsupport.ActionGetHrUs:    {},
		policytestsupport.ActionGetProfile: {},
	}
	newResourceRoles := policytestsupport.MakeRarList(newActionRoles)
	policyStoreSvc.expectUpdateResourceRoles(newResourceRoles[0], nil)
	policyStoreSvc.expectUpdateResourceRoles(newResourceRoles[1], nil)

	policies := policytestsupport.MakeRoleSubjectTestPolicies(newActionRoles)

	service := awsapigw.NewAwsApiGatewayProviderService(nil, policyStoreSvc)
	appInfo := orchestrator.ApplicationInfo{}
	status, err := service.SetPolicyInfo(appInfo, policies)
	assert.NoError(t, err)
	assert.Equal(t, http.StatusCreated, status)
}

func TestSetPolicyInfo_NoChange(t *testing.T) {
	policyStoreSvc := &mockPolicyStoreSvc{}
	existingActionRoles := map[string][]string{
		policytestsupport.ActionGetHrUs:    {"some-hr-role"},
		policytestsupport.ActionGetProfile: {"some-profile-role"},
	}
	expReturnResourceRoles := policytestsupport.MakeRarList(existingActionRoles)
	policyStoreSvc.expectGetResourceRoles(expReturnResourceRoles, nil)

	newActionRoles := map[string][]string{
		policytestsupport.ActionGetHrUs:    {"some-hr-role"},
		policytestsupport.ActionGetProfile: {"some-profile-role"},
	}
	policies := policytestsupport.MakeRoleSubjectTestPolicies(newActionRoles)

	service := awsapigw.NewAwsApiGatewayProviderService(nil, policyStoreSvc)
	appInfo := orchestrator.ApplicationInfo{}
	status, err := service.SetPolicyInfo(appInfo, policies)
	assert.NoError(t, err)
	assert.Equal(t, http.StatusCreated, status)
}

func TestSetPolicyInfo_AddsNewMembersAll(t *testing.T) {
	policyStoreSvc := &mockPolicyStoreSvc{}
	existingActionRoles := map[string][]string{
		policytestsupport.ActionGetHrUs:    {},
		policytestsupport.ActionGetProfile: {},
	}
	expReturnResourceRoles := policytestsupport.MakeRarList(existingActionRoles)
	policyStoreSvc.expectGetResourceRoles(expReturnResourceRoles, nil)

	newActionRoles := map[string][]string{
		policytestsupport.ActionGetHrUs:    {"some-hr-role"},
		policytestsupport.ActionGetProfile: {"some-profile-role"},
	}
	newResourceRoles := policytestsupport.MakeRarList(newActionRoles)
	policyStoreSvc.expectUpdateResourceRoles(newResourceRoles[0], nil)
	policyStoreSvc.expectUpdateResourceRoles(newResourceRoles[1], nil)
	policies := policytestsupport.MakeRoleSubjectTestPolicies(newActionRoles)

	service := awsapigw.NewAwsApiGatewayProviderService(nil, policyStoreSvc)
	appInfo := orchestrator.ApplicationInfo{}
	status, err := service.SetPolicyInfo(appInfo, policies)
	assert.NoError(t, err)
	assert.Equal(t, http.StatusCreated, status)
}

type mockPolicyStoreSvc struct {
	mock.Mock
}

func (m *mockPolicyStoreSvc) GetResourceRoles() ([]providerscommon.ResourceActionRoles, error) {
	args := m.Called()
	rars := args.Get(0)
	if rars == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]providerscommon.ResourceActionRoles), args.Error(1)
}

func (m *mockPolicyStoreSvc) UpdateResourceRole(rar providerscommon.ResourceActionRoles) error {
	args := m.Called(rar)
	return args.Error(0)
}

func (m *mockPolicyStoreSvc) expectGetResourceRoles(andReturn []providerscommon.ResourceActionRoles, orError error) {
	m.On("GetResourceRoles").Return(andReturn, orError)
}

func (m *mockPolicyStoreSvc) expectUpdateResourceRoles(withResRole providerscommon.ResourceActionRoles, orError error) {
	m.On("UpdateResourceRole", withResRole).Return(orError)
}
