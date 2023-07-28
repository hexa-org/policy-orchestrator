package dynamodbpolicy_test

import (
	"context"
	"github.com/aws/aws-sdk-go-v2/feature/dynamodb/attributevalue"
	ddb "github.com/aws/aws-sdk-go-v2/service/dynamodb"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb/types"
	"github.com/hexa-org/policy-orchestrator/internal/orchestratorproviders/amazonwebservices/awsapigw/dynamodbpolicy"
	"github.com/hexa-org/policy-orchestrator/internal/orchestratorproviders/amazonwebservices/awscommon"
	"github.com/hexa-org/policy-orchestrator/pkg/testsupport/awstestsupport"
	"github.com/stretchr/testify/assert"
	"net/http"
	"testing"
	"time"
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
	client := newDynamoDbClient()
	input := &ddb.ScanInput{TableName: &ddbTableName}
	out, err := client.Scan(context.TODO(), input)
	assert.ErrorContains(t, err, "StatusCode: 400")
	assert.Nil(t, out)
}

func TestUpdateItem(t *testing.T) {
	principalId, _ := attributevalue.Marshal("somePrincipal")
	resource, _ := attributevalue.Marshal("someResource")
	keyAttrVal := map[string]types.AttributeValue{"PrincipalId": principalId, "Resource": resource}
	input := &ddb.UpdateItemInput{TableName: &ddbTableName, Key: keyAttrVal}

	client := newDynamoDbClient()
	out, err := client.UpdateItem(context.TODO(), input)
	assert.ErrorContains(t, err, "StatusCode: 400")
	assert.Nil(t, out)
}

func newDynamoDbClient() dynamodbpolicy.DynamodbClient {
	httpClient := &http.Client{Timeout: time.Second}
	client, _ := dynamodbpolicy.NewDynamodbClient(awstestsupport.AwsCredentialsForTest(), awscommon.AWSClientOptions{DisableRetry: true, HTTPClient: httpClient})
	return client
}
