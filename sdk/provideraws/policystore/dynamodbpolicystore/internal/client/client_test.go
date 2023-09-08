package client_test

import (
	"context"
	"errors"
	awshttp "github.com/aws/aws-sdk-go-v2/aws/transport/http"
	ddb "github.com/aws/aws-sdk-go-v2/service/dynamodb"
	"github.com/aws/smithy-go"
	"github.com/hexa-org/policy-orchestrator/sdk/provideraws/policystore/dynamodbpolicystore/internal/client"
	"github.com/hexa-org/policy-orchestrator/sdk/provideraws/policystore/dynamodbpolicystore/internal/testhelper"
	"github.com/stretchr/testify/assert"
	"net/http"
	"testing"
)

func TestScan_Error(t *testing.T) {
	m := testhelper.NewMockDynamodbHttpClient()
	c, _ := client.NewDynamodbClient(testhelper.AwsCredentialsForTest(), m)
	tableName := testhelper.TableName
	input := &ddb.ScanInput{TableName: &tableName}

	m.ExpectScan(errors.New("some error"))
	output, err := c.Scan(context.TODO(), input)
	assert.ErrorContains(t, err, "some error")
	assert.Nil(t, output)
	opErr, respErr := awsError(err)
	assert.Equal(t, "DynamoDB", opErr.Service())
	assert.Equal(t, "Scan", opErr.Operation())
	assert.Equal(t, http.StatusBadRequest, respErr.HTTPStatusCode())
}

func TestScan(t *testing.T) {
	m := testhelper.NewMockDynamodbHttpClient()
	c, _ := client.NewDynamodbClient(testhelper.AwsCredentialsForTest(), m)
	tableName := testhelper.TableName
	input := &ddb.ScanInput{TableName: &tableName}
	testItem := testhelper.MakeResourceActionRoles()
	m.ExpectScan(nil, testItem)
	output, err := c.Scan(context.TODO(), input)
	assert.NoError(t, err)
	assert.NotNil(t, output)
	assert.Equal(t, testhelper.ScanOutput().Items, output.Items)
}

func TestUpdateItem_Error(t *testing.T) {
	m := testhelper.NewMockDynamodbHttpClient()
	c, _ := client.NewDynamodbClient(testhelper.AwsCredentialsForTest(), m)

	reqItem := testhelper.MakeResourceActionRoles()
	m.ExpectUpdateItem(reqItem, errors.New("some-error"))

	inputBuilder := inputBuilderV2()
	input, _ := inputBuilder.UpdateItemInput(reqItem)

	_, err := c.UpdateItem(context.TODO(), input)
	assert.ErrorContains(t, err, "some-error")
	opErr, respErr := awsError(err)
	assert.Equal(t, "DynamoDB", opErr.Service())
	assert.Equal(t, "UpdateItem", opErr.Operation())
	assert.Equal(t, http.StatusBadRequest, respErr.HTTPStatusCode())

}

func TestUpdateItem(t *testing.T) {
	m := testhelper.NewMockDynamodbHttpClient()
	c, _ := client.NewDynamodbClient(testhelper.AwsCredentialsForTest(), m)

	reqItem := testhelper.MakeResourceActionRoles()
	m.ExpectUpdateItem(reqItem, nil)

	inputBuilder := inputBuilderV2()
	input, _ := inputBuilder.UpdateItemInput(reqItem)

	_, err := c.UpdateItem(context.TODO(), input)
	assert.NoError(t, err)
}

func inputBuilderV2() *client.InputBuilderV2 {
	return client.NewInputBuilderV2(testhelper.TableName, testhelper.DynamicTableDefinition())
}

func awsError(err error) (*smithy.OperationError, *awshttp.ResponseError) {
	var opErr *smithy.OperationError
	_ = errors.As(err, &opErr)

	var respErr *awshttp.ResponseError
	_ = errors.As(opErr.Err, &respErr)
	return opErr, respErr
}
