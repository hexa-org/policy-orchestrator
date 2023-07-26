package dynamodbpolicy_test

import (
	"context"
	ddb "github.com/aws/aws-sdk-go-v2/service/dynamodb"
	"github.com/hexa-org/policy-orchestrator/internal/orchestratorproviders/amazonwebservices/awsapigw/dynamodbpolicy"
	"github.com/hexa-org/policy-orchestrator/internal/orchestratorproviders/amazonwebservices/awscommon"
	"github.com/hexa-org/policy-orchestrator/pkg/testsupport/awstestsupport"
	"github.com/stretchr/testify/assert"
	"testing"
)

var ddbTableName = "some-ddb-table"

func TestNewDynamodbClient_Error(t *testing.T) {
	client, err := dynamodbpolicy.NewDynamodbClient([]byte("a"), awscommon.AWSClientOptions{DisableRetry: true})
	assert.ErrorContains(t, err, "invalid character 'a'")
	assert.Nil(t, client)
}

func TestNewDynamodbClient_Success(t *testing.T) {
	client, err := dynamodbpolicy.NewDynamodbClient(awstestsupport.AwsCredentialsForTest(), awscommon.AWSClientOptions{DisableRetry: true})
	assert.NoError(t, err)
	assert.NotNil(t, client)
}

func TestScan(t *testing.T) {
	client, _ := dynamodbpolicy.NewDynamodbClient(awstestsupport.AwsCredentialsForTest(), awscommon.AWSClientOptions{DisableRetry: true})
	input := &ddb.ScanInput{TableName: &ddbTableName}
	out, err := client.Scan(context.TODO(), input)
	assert.ErrorContains(t, err, "StatusCode: 400")
	assert.Nil(t, out)
}

func TestUpdateItem(t *testing.T) {
	client, _ := dynamodbpolicy.NewDynamodbClient(awstestsupport.AwsCredentialsForTest(), awscommon.AWSClientOptions{DisableRetry: true})
	input := &ddb.UpdateItemInput{TableName: &ddbTableName}
	out, err := client.UpdateItem(context.TODO(), input)
	assert.Error(t, err)
	assert.Nil(t, out)
}
