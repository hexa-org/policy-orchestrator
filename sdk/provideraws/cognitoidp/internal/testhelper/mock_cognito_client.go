package testhelper

import (
	"github.com/aws/aws-sdk-go-v2/service/cognitoidentityprovider"
	"github.com/stretchr/testify/mock"
)

type MockCognitoClient struct {
	mock.Mock
}

func NewMockCognitoClient() *MockCognitoClient {
	return &MockCognitoClient{}
}

func (mc *MockCognitoClient) ListUserPools() (*cognitoidentityprovider.ListUserPoolsOutput, error) {
	args := mc.Called()
	return args.Get(0).(*cognitoidentityprovider.ListUserPoolsOutput), args.Error(1)
}

func (mc *MockCognitoClient) ListResourceServers(userPoolId string) (*cognitoidentityprovider.ListResourceServersOutput, error) {
	args := mc.Called(userPoolId)
	return args.Get(0).(*cognitoidentityprovider.ListResourceServersOutput), args.Error(1)
}

func (mc *MockCognitoClient) ExpectListUserPools(output *cognitoidentityprovider.ListUserPoolsOutput, err error) {
	mc.On("ListUserPools").Return(output, err)
}

func (mc *MockCognitoClient) ExpectListResourceServers(userPoolId string, withOutput *cognitoidentityprovider.ListResourceServersOutput, withErr error) {
	mc.On("ListResourceServers", userPoolId).Return(withOutput, withErr)
}
