package amazonwebservices_test

import (
	"context"
	"github.com/aws/aws-sdk-go-v2/service/cognitoidentityprovider"
	"github.com/aws/aws-sdk-go-v2/service/cognitoidentityprovider/types"
	"github.com/stretchr/testify/mock"
)

type MockClient struct {
	mock.Mock
	Err error
}

func (m *MockClient) ListUsers(ctx context.Context, params *cognitoidentityprovider.ListUsersInput, optFns ...func(*cognitoidentityprovider.Options)) (*cognitoidentityprovider.ListUsersOutput, error) {
	username := "aUser"
	name := "email"
	value := "aUser@amazon.com"
	attributes := []types.AttributeType{{Name: &name, Value: &value}}
	return &cognitoidentityprovider.ListUsersOutput{Users: []types.UserType{{Username: &username, Attributes: attributes}}}, m.Err
}

func (m *MockClient) ListUserPools(ctx context.Context, params *cognitoidentityprovider.ListUserPoolsInput, optFns ...func(*cognitoidentityprovider.Options)) (*cognitoidentityprovider.ListUserPoolsOutput, error) {
	id := "anId"
	name := "aName"
	return &cognitoidentityprovider.ListUserPoolsOutput{
		UserPools: []types.UserPoolDescriptionType{{Id: &id, Name: &name}},
	}, m.Err
}

