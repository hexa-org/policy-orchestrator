package cognitoidp

import (
	"context"
	"github.com/aws/aws-sdk-go-v2/service/cognitoidentityprovider"
	"github.com/hexa-org/policy-orchestrator/v2/provideraws/awscommon"
	logger "golang.org/x/exp/slog"
)

type CognitoClient interface {
	listUserPools() (*cognitoidentityprovider.ListUserPoolsOutput, error)
	listResourceServers(userPoolId string) (*cognitoidentityprovider.ListResourceServersOutput, error)
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

func (c *cognitoClient) listUserPools() (*cognitoidentityprovider.ListUserPoolsOutput, error) {
	poolsInput := cognitoidentityprovider.ListUserPoolsInput{MaxResults: 20}
	pools, err := c.client.ListUserPools(context.Background(), &poolsInput)
	return pools, err
}

func (c *cognitoClient) listResourceServers(userPoolId string) (*cognitoidentityprovider.ListResourceServersOutput, error) {
	rsInput := cognitoidentityprovider.ListResourceServersInput{UserPoolId: &userPoolId, MaxResults: 10}
	rsOutput, err := c.client.ListResourceServers(context.Background(), &rsInput)
	return rsOutput, err
}
