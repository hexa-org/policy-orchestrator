package dynamodbpolicy

import (
	"context"
	ddb "github.com/aws/aws-sdk-go-v2/service/dynamodb"
	"github.com/hexa-org/policy-orchestrator/demo/internal/orchestratorproviders/amazonwebservices/awscommon"
	log "golang.org/x/exp/slog"
)

type DynamodbClient interface {
	Scan(ctx context.Context, params *ddb.ScanInput, optFns ...func(*ddb.Options)) (*ddb.ScanOutput, error)
	UpdateItem(ctx context.Context, params *ddb.UpdateItemInput, optFns ...func(*ddb.Options)) (*ddb.UpdateItemOutput, error)
}

type dynamodbClient struct {
	internal *ddb.Client
}

func NewDynamodbClient(key []byte, opt awscommon.AWSClientOptions) (DynamodbClient, error) {
	cfg, err := awscommon.GetAwsClientConfig(key, opt)
	if err != nil {
		log.Error("NewDynamodbClient Failed to GetAwsClientConfig", "Error", err)
		return nil, err
	}
	internalClient := ddb.NewFromConfig(cfg)
	return &dynamodbClient{internal: internalClient}, nil
}

func (c *dynamodbClient) Scan(ctx context.Context, params *ddb.ScanInput, optFns ...func(*ddb.Options)) (*ddb.ScanOutput, error) {
	return c.internal.Scan(ctx, params, optFns...)
}

func (c *dynamodbClient) UpdateItem(ctx context.Context, params *ddb.UpdateItemInput, optFns ...func(*ddb.Options)) (*ddb.UpdateItemOutput, error) {
	return c.internal.UpdateItem(ctx, params, optFns...)
}
