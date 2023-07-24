package cognitotestsupport

import (
	"encoding/json"
	"fmt"
	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/cognitoidentityprovider"
	"github.com/aws/aws-sdk-go-v2/service/cognitoidentityprovider/types"
	"github.com/hexa-org/policy-orchestrator/internal/orchestrator"
	"net/http"
)

const TestAwsRegion = "aRegion"
const TestAwsAccessKeyId = "anAccessKeyID"
const TestAwsSecretAccessKey = "aSecretAccessKey"

const TestUserPoolId = "some-user-pool-id"
const TestUserPoolName = "some-user-pool-name"
const TestResourceServerIdentifier = "https://some-resource-server"
const TestResourceServerName = "some-resource-server-name"

var CognitoApiUrl = fmt.Sprintf("https://cognito-idp.%s.amazonaws.com/", TestAwsRegion)

func AwsCredentialsForTest() []byte {
	str := fmt.Sprintf(`
{
  "accessKeyID": "%s",
  "secretAccessKey": "%s",
  "region": "%s"
}
`, TestAwsAccessKeyId, TestAwsSecretAccessKey, TestAwsRegion)

	return []byte(str)
}

func IntegrationInfo() orchestrator.IntegrationInfo {
	return orchestrator.IntegrationInfo{Name: "amazon", Key: AwsCredentialsForTest()}
}

func AppInfo() orchestrator.ApplicationInfo {
	return orchestrator.ApplicationInfo{
		ObjectID:    TestUserPoolId,
		Name:        TestResourceServerName,
		Description: "Cognito",
		Service:     TestResourceServerIdentifier,
	}
}

func (m *MockCognitoHTTPClient) MockListUserPools() {
	m.MockListUserPoolsWithHttpStatus(http.StatusOK)
}

func (m *MockCognitoHTTPClient) MockListUserPoolsWithHttpStatus(httpStatus int) {
	m.AddRequest(http.MethodPost, CognitoApiUrl, "ListUserPools", httpStatus, ListUserPoolsResponse())
}

func (m *MockCognitoHTTPClient) MockListResourceServers(withResp cognitoidentityprovider.ListResourceServersOutput) {
	m.MockListResourceServersWithHttpStatus(http.StatusOK, withResp)
}

func (m *MockCognitoHTTPClient) MockListResourceServersWithHttpStatus(httpStatus int, withResp cognitoidentityprovider.ListResourceServersOutput) {
	resp, _ := json.Marshal(withResp)
	m.AddRequest(http.MethodPost, CognitoApiUrl, "ListResourceServers", httpStatus, resp)
}

func (m *MockCognitoHTTPClient) MockListGroups(groupNames ...string) {
	m.AddRequest(http.MethodPost, CognitoApiUrl, "ListGroups", http.StatusOK, ListGroupsResponse(groupNames...))
}

func (m *MockCognitoHTTPClient) MockListGroupsWithHttpStatus(httpStatus int, groupNames ...string) {
	m.AddRequest(http.MethodPost, CognitoApiUrl, "ListGroups", httpStatus, ListGroupsResponse(groupNames...))
}

func (m *MockCognitoHTTPClient) MockListUsersInGroup(userName ...string) {
	m.MockListUsersInGroupWithHttpStatus(http.StatusOK, userName...)
}

func (m *MockCognitoHTTPClient) MockListUsersInGroupWithHttpStatus(httpStatus int, userName ...string) {
	usersInGroupResp := ListUsersInGroupResponse(userName...)
	m.AddRequest(http.MethodPost, CognitoApiUrl, "ListUsersInGroup", httpStatus, usersInGroupResp)
}

func (m *MockCognitoHTTPClient) MockAdminGetUser(userName, email string) {
	adminGetUserResp := AdminGetUserResponse(userName, email)
	m.AddRequest(http.MethodPost, CognitoApiUrl, "AdminGetUser", http.StatusOK, adminGetUserResp)
}

func (m *MockCognitoHTTPClient) MockListUsers(principalId string) {
	listUsersResp := ListUsersResponse(principalId)
	m.AddRequest(http.MethodPost, CognitoApiUrl, "ListUsers", http.StatusOK, listUsersResp)
}

func (m *MockCognitoHTTPClient) MockAdminAddUserToGroup() {
	m.MockAdminAddUserToGroupWithHttpStatus(http.StatusOK)
}

func (m *MockCognitoHTTPClient) MockAdminAddUserToGroupWithHttpStatus(httpStatus int) {
	addUsersToGroupResp := AdminAddUserToGroupResponse()
	m.AddRequest(http.MethodPost, CognitoApiUrl, "AdminAddUserToGroup", httpStatus, addUsersToGroupResp)
}

func (m *MockCognitoHTTPClient) MockAdminRemoveUserFromGroup() {
	m.MockAdminRemoveUserFromGroupWithHttpStatus(http.StatusOK)
}

func (m *MockCognitoHTTPClient) MockAdminRemoveUserFromGroupWithHttpStatus(httpStatus int) {
	removeResp := AdminRemoveUserFromGroupResponse()
	m.AddRequest(http.MethodPost, CognitoApiUrl, "AdminRemoveUserFromGroup", httpStatus, removeResp)
}

func ListGroupsResponse(groupNames ...string) []byte {
	groups := make([]types.GroupType, 0)
	for _, name := range groupNames {
		groups = append(groups, types.GroupType{
			GroupName:   aws.String(name),
			UserPoolId:  aws.String(TestUserPoolId),
			Description: aws.String("some description")})
	}
	output := cognitoidentityprovider.ListGroupsOutput{Groups: groups}
	expBytes, _ := json.Marshal(output)
	return expBytes
}

func ListUsersInGroupResponse(principalIds ...string) []byte {
	users := make([]types.UserType, 0)
	for _, username := range principalIds {
		users = append(users, types.UserType{
			Username: aws.String(username)})
	}
	output := cognitoidentityprovider.ListUsersInGroupOutput{Users: users}
	expBytes, _ := json.Marshal(output)
	return expBytes
}

func AdminGetUserResponse(principalId, email string) []byte {
	attrs := []types.AttributeType{
		{
			Name:  aws.String("email"),
			Value: aws.String(email),
		},
	}

	output := cognitoidentityprovider.AdminGetUserOutput{
		Username:       aws.String(principalId),
		UserAttributes: attrs}
	expBytes, _ := json.Marshal(output)
	return expBytes
}

func ListUsersResponse(principalId string) []byte {

	var users []types.UserType
	if principalId != "" {
		users = []types.UserType{{Username: &principalId}}
	}

	output := cognitoidentityprovider.ListUsersOutput{
		Users: users,
	}
	expBytes, _ := json.Marshal(output)
	return expBytes
}

func AdminAddUserToGroupResponse() []byte {
	output := cognitoidentityprovider.AdminAddUserToGroupOutput{}
	expBytes, _ := json.Marshal(output)
	return expBytes
}

func AdminRemoveUserFromGroupResponse() []byte {
	output := cognitoidentityprovider.AdminRemoveUserFromGroupOutput{}
	expBytes, _ := json.Marshal(output)
	return expBytes
}

func ListUserPoolsResponse() []byte {
	listOutput := cognitoidentityprovider.ListUserPoolsOutput{
		NextToken: nil,
		UserPools: []types.UserPoolDescriptionType{
			{
				Id:   aws.String(TestUserPoolId),
				Name: aws.String(TestUserPoolName),
			},
		},
	}

	expBytes, _ := json.Marshal(listOutput)
	return expBytes
}

func WithResourceServer() cognitoidentityprovider.ListResourceServersOutput {
	return WithResourceServerOptions(TestUserPoolId, TestResourceServerName, TestResourceServerIdentifier)
}

func WithResourceServerOptions(userPoolId, name, identifier string) cognitoidentityprovider.ListResourceServersOutput {
	usePoolId := TestUserPoolId
	useName := TestResourceServerName
	useIdentifier := TestResourceServerIdentifier

	if userPoolId != "" {
		usePoolId = userPoolId
	}
	if name != "" {
		useName = name
	}
	if identifier != "" {
		useIdentifier = identifier
	}

	return cognitoidentityprovider.ListResourceServersOutput{
		ResourceServers: []types.ResourceServerType{
			{
				Identifier: &useIdentifier,
				Name:       &useName,
				UserPoolId: &usePoolId,
			},
		},
	}
}
