package client

import (
	"context"

	"github.com/aws/aws-sdk-go-v2/service/cognitoidentityprovider"
	"github.com/hexa-org/policy-orchestrator/sdk/provideraws/awscommon"
	logger "golang.org/x/exp/slog"
)

type CognitoClient interface {
	ListUserPools() (*cognitoidentityprovider.ListUserPoolsOutput, error)
	ListResourceServers(userPoolId string) (*cognitoidentityprovider.ListResourceServersOutput, error)
}

type cognitoClient struct {
	client *cognitoidentityprovider.Client
}

func NewCognitoClient(key []byte, httpClient awscommon.AWSHttpClient) (CognitoClient, error) {
	cfg, err := awscommon.GetAwsClientConfig(key, httpClient)
	if err != nil {
		logger.Error("NewCognitoClient", "error building aws client config", "error", err.Error())
		return nil, err
	}
	return &cognitoClient{client: cognitoidentityprovider.NewFromConfig(cfg)}, nil
}

func (c *cognitoClient) ListUserPools() (*cognitoidentityprovider.ListUserPoolsOutput, error) {
	maxRes := int32(20)
	poolsInput := cognitoidentityprovider.ListUserPoolsInput{MaxResults: &maxRes}
	pools, err := c.client.ListUserPools(context.Background(), &poolsInput)
	return pools, err
}

func (c *cognitoClient) ListResourceServers(userPoolId string) (*cognitoidentityprovider.ListResourceServersOutput, error) {
	maxRes := int32(10)
	rsInput := cognitoidentityprovider.ListResourceServersInput{UserPoolId: &userPoolId, MaxResults: &maxRes}
	rsOutput, err := c.client.ListResourceServers(context.Background(), &rsInput)
	return rsOutput, err
}
