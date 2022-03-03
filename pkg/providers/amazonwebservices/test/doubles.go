package amazonwebservices_test

import (
	"context"
	"github.com/aws/aws-sdk-go-v2/service/cognitoidentityprovider"
	"github.com/aws/aws-sdk-go-v2/service/cognitoidentityprovider/types"
	"github.com/stretchr/testify/mock"
)

type MockClient struct {
	mock.Mock
	Errs map[string]error
}

func (m *MockClient) ListUsers(_ context.Context, _ *cognitoidentityprovider.ListUsersInput, _ ...func(*cognitoidentityprovider.Options)) (*cognitoidentityprovider.ListUsersOutput, error) {
	username := "aUser"
	name := "email"
	value := "aUser@amazon.com"
	attributes := []types.AttributeType{{Name: &name, Value: &value}}
	return &cognitoidentityprovider.ListUsersOutput{
		Users: []types.UserType{{Username: &username, Attributes: attributes}},
	}, m.Errs["ListUsers"]
}

func (m *MockClient) ListUserPools(_ context.Context, _ *cognitoidentityprovider.ListUserPoolsInput, _ ...func(*cognitoidentityprovider.Options)) (*cognitoidentityprovider.ListUserPoolsOutput, error) {
	id := "anId"
	name := "aName"
	return &cognitoidentityprovider.ListUserPoolsOutput{
		UserPools: []types.UserPoolDescriptionType{{Id: &id, Name: &name}},
	}, m.Errs["ListUserPools"]
}

func (m *MockClient) AdminEnableUser(_ context.Context, _ *cognitoidentityprovider.AdminEnableUserInput, _ ...func(*cognitoidentityprovider.Options)) (*cognitoidentityprovider.AdminEnableUserOutput, error) {
	return &cognitoidentityprovider.AdminEnableUserOutput{}, m.Errs["AdminEnableUser"]
}

func (m *MockClient) AdminDisableUser(_ context.Context, _ *cognitoidentityprovider.AdminDisableUserInput, _ ...func(*cognitoidentityprovider.Options)) (*cognitoidentityprovider.AdminDisableUserOutput, error) {
	return &cognitoidentityprovider.AdminDisableUserOutput{}, m.Errs["AdminDisableUser"]
}
