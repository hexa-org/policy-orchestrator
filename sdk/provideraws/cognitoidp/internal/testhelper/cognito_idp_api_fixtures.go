package testhelper

import (
	"encoding/json"
	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/cognitoidentityprovider"
	"github.com/aws/aws-sdk-go-v2/service/cognitoidentityprovider/types"
)

// ListUserPoolsResponse - builds valid response body bytes for a
// successful listUserPools response
func ListUserPoolsResponse() []byte {
	expBytes, _ := json.Marshal(ListUserPoolsOutput())
	return expBytes
}

// ListUserPoolsOutput - builds a valid cognito ListUserPoolsOutput struct
// with the UserPoolId and UserPoolName
func ListUserPoolsOutput() *cognitoidentityprovider.ListUserPoolsOutput {
	return &cognitoidentityprovider.ListUserPoolsOutput{
		NextToken: nil,
		UserPools: []types.UserPoolDescriptionType{
			{
				Id:   aws.String(TestUserPoolId),
				Name: aws.String(TestUserPoolName),
			},
		},
	}
}

func ListResourceServersOutput() *cognitoidentityprovider.ListResourceServersOutput {
	return ListResourceServersOutputCustom(TestUserPoolId, TestResourceServerName, TestResourceServerIdentifier)
}

func ListResourceServersOutputCustom(userPoolId, name, identifier string) *cognitoidentityprovider.ListResourceServersOutput {
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

	return &cognitoidentityprovider.ListResourceServersOutput{
		ResourceServers: []types.ResourceServerType{
			{
				Identifier: &useIdentifier,
				Name:       &useName,
				UserPoolId: &usePoolId,
			},
		},
	}
}
