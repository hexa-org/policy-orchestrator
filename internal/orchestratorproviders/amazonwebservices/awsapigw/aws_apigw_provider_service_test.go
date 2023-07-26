package awsapigw_test

import (
	"errors"
	"github.com/hexa-org/policy-orchestrator/internal/orchestrator"
	"github.com/hexa-org/policy-orchestrator/internal/orchestratorproviders/amazonwebservices/awsapigw"
	"github.com/hexa-org/policy-orchestrator/internal/orchestratorproviders/providerscommon"
	"github.com/hexa-org/policy-orchestrator/pkg/testsupport/awstestsupport"
	"github.com/hexa-org/policy-orchestrator/pkg/testsupport/policytestsupport"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
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
	return args.Error(1)
}

func (m *mockPolicyStoreSvc) expectGetResourceRoles(andReturn []providerscommon.ResourceActionRoles, orError error) {
	m.On("GetResourceRoles").Return(andReturn, orError)
}
