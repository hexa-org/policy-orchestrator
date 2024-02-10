package client

import (
	"context"
	ddb "github.com/aws/aws-sdk-go-v2/service/dynamodb"
	"github.com/hexa-org/policy-orchestrator/sdk/provideraws/awscommon"
)

// DynamodbClient -
// BEGIN - copied from dynamodb_client.go
type DynamodbClient interface {
	Scan(ctx context.Context, params *ddb.ScanInput, optFns ...func(*ddb.Options)) (*ddb.ScanOutput, error)
	UpdateItem(ctx context.Context, params *ddb.UpdateItemInput, optFns ...func(*ddb.Options)) (*ddb.UpdateItemOutput, error)
}

type dynamodbClient struct {
	internal *ddb.Client
}

// NewDynamodbClient - builds DynamodbClient with provide credentials and optional httpClient
// pass an httpClient to use for tests
func NewDynamodbClient(key []byte, httpClient awscommon.AWSHttpClient) (DynamodbClient, error) {
	cfg, err := awscommon.GetAwsClientConfig(key, httpClient)
	if err != nil {
		return nil, err
	}

	return &dynamodbClient{internal: ddb.NewFromConfig(cfg)}, nil
}

func (c *dynamodbClient) Scan(ctx context.Context, params *ddb.ScanInput, optFns ...func(*ddb.Options)) (*ddb.ScanOutput, error) {
	out, err := c.internal.Scan(ctx, params, optFns...)
	return out, err
}

func (c *dynamodbClient) UpdateItem(ctx context.Context, params *ddb.UpdateItemInput, optFns ...func(*ddb.Options)) (*ddb.UpdateItemOutput, error) {
	return c.internal.UpdateItem(ctx, params, optFns...)
}

// END copied from dynamodb_client.go
