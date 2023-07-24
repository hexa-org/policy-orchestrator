package awsapigw_test

import (
	"errors"
	"github.com/hexa-org/policy-orchestrator/internal/orchestrator"
	"github.com/hexa-org/policy-orchestrator/internal/orchestratorproviders/amazonwebservices/awsapigw"
	"github.com/hexa-org/policy-orchestrator/pkg/testsupport/cognitotestsupport"
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
	service := awsapigw.NewAwsApiGatewayProviderService(cognitoClient)
	apps, err := service.DiscoverApplications(cognitotestsupport.IntegrationInfo())
	assert.Error(t, err)
	assert.Len(t, apps, 0)
}

func TestAwsApiGatewayProviderService_DiscoverApplications(t *testing.T) {
	cognitoClient := &mockCognitoClient{}
	expApps := []orchestrator.ApplicationInfo{cognitotestsupport.AppInfo()}
	cognitoClient.expectListUserPools(expApps, nil)
	service := awsapigw.NewAwsApiGatewayProviderService(cognitoClient)
	apps, err := service.DiscoverApplications(cognitotestsupport.IntegrationInfo())
	assert.NoError(t, err)
	assert.Len(t, apps, len(expApps))
}
