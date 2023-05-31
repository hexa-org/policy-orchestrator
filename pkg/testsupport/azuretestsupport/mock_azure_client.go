package azuretestsupport

import (
	"github.com/hexa-org/policy-orchestrator/internal/orchestrator"
	"github.com/hexa-org/policy-orchestrator/internal/orchestratorproviders/microsoftazure"
	"github.com/hexa-org/policy-orchestrator/pkg/testsupport/policytestsupport"
	"github.com/stretchr/testify/mock"
	"reflect"
)

type MockAzureClient struct {
	mock.Mock
}

func (m *MockAzureClient) GetWebApplications(key []byte) ([]orchestrator.ApplicationInfo, error) {
	returnArgs := m.Called(key)
	return returnArgs.Get(0).([]orchestrator.ApplicationInfo), returnArgs.Error(1)
}

func (m *MockAzureClient) GetServicePrincipals(key []byte, appId string) (microsoftazure.AzureServicePrincipals, error) {
	returnArgs := m.Called(key, appId)
	return returnArgs.Get(0).(microsoftazure.AzureServicePrincipals), returnArgs.Error(1)
}

func (m *MockAzureClient) GetUserInfoFromPrincipalId(key []byte, principalId string) (microsoftazure.AzureUser, error) {
	returnArgs := m.Called(key, principalId)
	return returnArgs.Get(0).(microsoftazure.AzureUser), returnArgs.Error(1)
}

func (m *MockAzureClient) GetPrincipalIdFromEmail(key []byte, email string) (string, error) {
	returnArgs := m.Called(key, email)
	return returnArgs.String(0), returnArgs.Error(1)
}

func (m *MockAzureClient) GetAppRoleAssignedTo(key []byte, servicePrincipalId string) (microsoftazure.AzureAppRoleAssignments, error) {
	returnArgs := m.Called(key, servicePrincipalId)
	return returnArgs.Get(0).(microsoftazure.AzureAppRoleAssignments), returnArgs.Error(1)
}

func (m *MockAzureClient) SetAppRoleAssignedTo(key []byte, servicePrincipalId string, assignments []microsoftazure.AzureAppRoleAssignment) error {
	returnArgs := m.Called(key, servicePrincipalId, assignments)
	return returnArgs.Error(0)
}

func NewMockAzureClient() *MockAzureClient {
	return &MockAzureClient{}
}

func (m *MockAzureClient) ExpectGetServicePrincipals() {
	m.On("GetServicePrincipals", AzureClientKey(), AzureAppId).
		Return(AzureServicePrincipals(), nil)
}

func (m *MockAzureClient) ExpectAppRoleAssignedTo(assignments []microsoftazure.AzureAppRoleAssignment) {
	m.On("GetAppRoleAssignedTo", AzureClientKey(), ServicePrincipalId).
		Return(MakeAssignments(assignments), nil)
}

func (m *MockAzureClient) ExpectGetUserInfoFromPrincipalId(principalIds ...string) {
	for _, pId := range principalIds {
		m.On("GetUserInfoFromPrincipalId", AzureClientKey(), pId).
			Return(microsoftazure.AzureUser{
				PrincipalId: pId,
				Email:       policytestsupport.MakeEmail(pId),
			}, nil)
	}
}

func (m *MockAzureClient) ExpectGetPrincipalIdFromEmail(email, principalId string) {
	m.On("GetPrincipalIdFromEmail", AzureClientKey(), email).
		Return(principalId, nil)
}

func (m *MockAzureClient) ExpectSetAppRoleAssignedTo(requestedAssignments []microsoftazure.AzureAppRoleAssignment) {
	theFunc := mock.MatchedBy(func(actAssignments []microsoftazure.AzureAppRoleAssignment) bool {
		if len(actAssignments) != len(requestedAssignments) {
			return false
		}

		expSorted := SortAssignments(AssignmentsWithoutId(requestedAssignments))
		actSorted := SortAssignments(AssignmentsWithoutId(actAssignments))

		return reflect.DeepEqual(expSorted, actSorted)
	})

	m.On("SetAppRoleAssignedTo", AzureClientKey(), ServicePrincipalId, theFunc).
		Return(nil)
}
